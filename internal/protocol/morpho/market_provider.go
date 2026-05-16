package morpho

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"test-task-edge/internal/types"
)

type marketScanner interface {
	ScanMarkets(ctx context.Context, blockNumber uint64) ([]types.Bytes32, error)
}

type MarketProvider struct {
	scanner marketScanner
	cache   sync.Map
}

func NewMarketProvider(scanner marketScanner) *MarketProvider {
	return &MarketProvider{scanner: scanner}
}

func (p *MarketProvider) DiscoverMarkets(ctx context.Context, blockNumber uint64) ([]types.Bytes32, error) {
	newIDs, err := p.scanner.ScanMarkets(ctx, blockNumber)
	if err != nil {
		return nil, err
	}

	for _, id := range newIDs {
		p.cache.Store(id.String(), id)
	}

	var all []types.Bytes32
	p.cache.Range(func(key, value interface{}) bool {
		all = append(all, value.(types.Bytes32))
		return true
	})

	if len(newIDs) > 0 {
		log.Info().Int("new", len(newIDs)).Int("total", len(all)).Msg("discovered morpho markets")
	}

	return all, nil
}
