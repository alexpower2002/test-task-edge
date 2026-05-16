package ethereum

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog/log"
	"test-task-edge/internal/types"
)

type gethCaller interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*gethTypes.Header, error)
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]gethTypes.Log, error)
	Close()
}

type RPCClient struct {
	ethCaller    gethCaller
	pollInterval time.Duration
	MaxRetries   int
	RetryWait    time.Duration
	RPCTimeout   time.Duration
}

func NewRPCClient(url string, pollInterval time.Duration) (*RPCClient, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
		TLSNextProto:    make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:    2,
		IdleConnTimeout: 30 * time.Second,
	}
	httpClient := &http.Client{Transport: tr, Timeout: 0}

	rpcClient, err := rpc.DialOptions(context.Background(), url, rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	ethClient := ethclient.NewClient(rpcClient)
	return &RPCClient{
		ethCaller:    ethClient,
		pollInterval: pollInterval,
		MaxRetries:   3,
		RetryWait:    time.Second,
		RPCTimeout:   15 * time.Second,
	}, nil
}

func NewRPCClientWithClient(ethCaller gethCaller, pollInterval time.Duration) *RPCClient {
	return &RPCClient{
		ethCaller:    ethCaller,
		pollInterval: pollInterval,
		MaxRetries:   3,
		RetryWait:    time.Second,
		RPCTimeout:   15 * time.Second,
	}
}

func (c *RPCClient) callWithRetry(ctx context.Context, label string, fn func(ctx context.Context) error) error {
	var lastErr error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		if attempt > 0 {
			wait := c.RetryWait << (attempt - 1)
			log.Warn().Err(lastErr).Str("rpc_call", label).Int("attempt", attempt).Dur("wait", wait).Msg("retrying RPC call")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}
		callCtx, cancel := context.WithTimeout(ctx, c.RPCTimeout)
		err := fn(callCtx)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
	}
	return lastErr
}

func (c *RPCClient) Close() {
	c.ethCaller.Close()
}

func (c *RPCClient) PollInterval() time.Duration {
	return c.pollInterval
}

func (c *RPCClient) LatestBlock(ctx context.Context) (types.BlockRef, error) {
	var header *gethTypes.Header
	err := c.callWithRetry(ctx, "LatestBlock", func(ctx context.Context) error {
		var err error
		header, err = c.ethCaller.HeaderByNumber(ctx, nil)
		return err
	})
	if err != nil {
		return types.BlockRef{}, err
	}
	return types.BlockRef{Number: header.Number.Uint64(), Timestamp: header.Time}, nil
}

func (c *RPCClient) BlockByNumber(ctx context.Context, number uint64) (types.BlockRef, error) {
	var header *gethTypes.Header
	err := c.callWithRetry(ctx, "BlockByNumber", func(ctx context.Context) error {
		var err error
		header, err = c.ethCaller.HeaderByNumber(ctx, new(big.Int).SetUint64(number))
		return err
	})
	if err != nil {
		return types.BlockRef{}, err
	}
	return types.BlockRef{Number: header.Number.Uint64(), Timestamp: header.Time}, nil
}

func (c *RPCClient) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]gethTypes.Log, error) {
	var logs []gethTypes.Log
	err := c.callWithRetry(ctx, "FilterLogs", func(ctx context.Context) error {
		var err error
		logs, err = c.ethCaller.FilterLogs(ctx, query)
		return err
	})
	return logs, err
}

func (c *RPCClient) EthCall(ctx context.Context, to types.Address, data string, block uint64) (string, error) {
	msg := ethereum.CallMsg{
		To:   (*common.Address)(&to),
		Data: common.FromHex(data),
	}
	var blockNum *big.Int
	if block > 0 {
		blockNum = new(big.Int).SetUint64(block)
	}
	var result []byte
	err := c.callWithRetry(ctx, "EthCall", func(ctx context.Context) error {
		var err error
		result, err = c.ethCaller.CallContract(ctx, msg, blockNum)
		return err
	})
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(result), nil
}

type EthLog struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	BlockNumber string   `json:"blockNumber"`
	TxHash      string   `json:"transactionHash"`
	LogIndex    string   `json:"logIndex"`
}

func (c *RPCClient) Logs(ctx context.Context, address types.Address, from, to uint64, topics []string) ([]EthLog, error) {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(from),
		ToBlock:   new(big.Int).SetUint64(to),
		Addresses: []common.Address{common.Address(address)},
	}
	if len(topics) > 0 {
		hashes := make([]common.Hash, len(topics))
		for i, t := range topics {
			hashes[i] = common.HexToHash(t)
		}
		query.Topics = [][]common.Hash{hashes}
	}
	var logs []gethTypes.Log
	err := c.callWithRetry(ctx, "FilterLogs", func(ctx context.Context) error {
		var err error
		logs, err = c.ethCaller.FilterLogs(ctx, query)
		return err
	})
	if err != nil {
		return nil, err
	}
	result := make([]EthLog, len(logs))
	for i, l := range logs {
		topics := make([]string, len(l.Topics))
		for j, t := range l.Topics {
			topics[j] = t.Hex()
		}
		result[i] = EthLog{
			Address:     l.Address.Hex(),
			Topics:      topics,
			Data:        "0x" + hex.EncodeToString(l.Data),
			BlockNumber: fmt.Sprintf("0x%x", l.BlockNumber),
			TxHash:      l.TxHash.Hex(),
			LogIndex:    fmt.Sprintf("0x%x", l.Index),
		}
	}
	return result, nil
}
