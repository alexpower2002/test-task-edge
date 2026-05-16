package morpho

import (
	"context"
	"fmt"
	"math/big"
	"sync"

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
	mu          sync.Mutex
	scannedUpTo uint64
}

func NewMarketDiscoverer(logFilterer logFilterer, address types.Address, deployBlock uint64) *MarketDiscoverer {
	return &MarketDiscoverer{
		logFilterer: logFilterer,
		address:     address,
		scannedUpTo: deployBlock,
	}
}

func (d *MarketDiscoverer) ScanMarkets(ctx context.Context, blockNumber uint64) ([]types.Bytes32, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if blockNumber <= d.scannedUpTo {
		return nil, nil
	}

	from := d.scannedUpTo
	d.scannedUpTo = blockNumber

	var ids []types.Bytes32
	seen := make(map[string]bool)
	topic := common.HexToHash(createMarketTopic)

	for from <= blockNumber {
		to := from + scanBatchSize - 1
		if to > blockNumber {
			to = blockNumber
		}

		logs, err := d.logFilterer.FilterLogs(ctx, ethereum.FilterQuery{
			Addresses: []common.Address{common.Address(d.address)},
			Topics:    [][]common.Hash{{topic}},
			FromBlock: new(big.Int).SetUint64(from),
			ToBlock:   new(big.Int).SetUint64(to),
		})
		if err != nil {
			return nil, fmt.Errorf("scan markets block %d-%d: %w", from, to, err)
		}

		for _, l := range logs {
			if len(l.Topics) < 2 {
				continue
			}
			key := l.Topics[1].Hex()
			if seen[key] {
				continue
			}
			seen[key] = true
			var b types.Bytes32
			copy(b[:], l.Topics[1].Bytes())
			ids = append(ids, b)
		}

		from = to + 1
	}

	return ids, nil
}
