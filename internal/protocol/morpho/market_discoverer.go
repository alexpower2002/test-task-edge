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

type logFilterer interface {
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]gethTypes.Log, error)
}

type MarketDiscoverer struct {
	logFilterer logFilterer
	address     types.Address
}

func NewMarketDiscoverer(logFilterer logFilterer, address types.Address) *MarketDiscoverer {
	return &MarketDiscoverer{logFilterer: logFilterer, address: address}
}

func (d *MarketDiscoverer) ScanMarkets(ctx context.Context, blockNumber uint64) ([]types.Bytes32, error) {
	logs, err := d.logFilterer.FilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{common.Address(d.address)},
		FromBlock: new(big.Int).SetUint64(blockNumber),
		ToBlock:   new(big.Int).SetUint64(blockNumber),
	})
	if err != nil {
		return nil, fmt.Errorf("scan markets at block %d: %w", blockNumber, err)
	}

	seen := make(map[string]bool)
	var ids []types.Bytes32
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

	return ids, nil
}
