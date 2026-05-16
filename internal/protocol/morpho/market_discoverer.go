package morpho

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"test-task-edge/internal/types"
)

const (
	createMarketTopic = "0xac4b2400f169220b0c0afdde7a0b32e775ba727ea1cb30b35f935cdaab8683ac"
	scanBatchSize     = 10000
)

type logFilterer interface {
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]gethTypes.Log, error)
}

type MarketDiscoverer struct {
	logFilterer logFilterer
	address     types.Address
	fromBlock   uint64
}

func NewMarketDiscoverer(logFilterer logFilterer, address types.Address, fromBlock uint64) *MarketDiscoverer {
	return &MarketDiscoverer{logFilterer: logFilterer, address: address, fromBlock: fromBlock}
}

func (d *MarketDiscoverer) ScanMarkets(ctx context.Context, toBlock uint64) ([]types.Bytes32, error) {
	var ids []types.Bytes32
	seen := make(map[string]bool)

	topic := common.HexToHash(createMarketTopic)

	for from := d.fromBlock; from <= toBlock; from += scanBatchSize {
		to := from + scanBatchSize - 1
		if to > toBlock {
			to = toBlock
		}

		logs, err := d.logFilterer.FilterLogs(ctx, ethereum.FilterQuery{
			Addresses: []common.Address{common.Address(d.address)},
			FromBlock: new(big.Int).SetUint64(from),
			ToBlock:   new(big.Int).SetUint64(to),
			Topics:    [][]common.Hash{{topic}},
		})
		if err != nil {
			return nil, fmt.Errorf("scan markets at %d-%d: %w", from, to, err)
		}

		for _, l := range logs {
			if len(l.Topics) >= 2 {
				id := l.Topics[1].Hex()
				if !seen[id] {
					seen[id] = true
					var b types.Bytes32
					copy(b[:], l.Topics[1].Bytes())
					ids = append(ids, b)
				}
			}
		}
	}

	return ids, nil
}
