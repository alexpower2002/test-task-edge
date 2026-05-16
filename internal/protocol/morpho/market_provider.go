package morpho

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"test-task-edge/internal/types"
)

type marketScanner interface {
	ScanMarkets(ctx context.Context, latestBlock uint64) ([]types.Bytes32, error)
}

type MarketProvider struct {
	scanner marketScanner
	cache   sync.Map
}

var marketKey = "markets"

func NewMarketProvider(scanner marketScanner) *MarketProvider {
	return &MarketProvider{scanner: scanner}
}

func (p *MarketProvider) DiscoverMarkets(ctx context.Context, latestBlock uint64) ([]types.Bytes32, error) {
	if cached, ok := p.cache.Load(marketKey); ok {
		return cached.([]types.Bytes32), nil
	}

	ids, err := p.scanner.ScanMarkets(ctx, latestBlock)
	if err != nil {
		return nil, err
	}

	p.cache.Store(marketKey, ids)

	log.Info().Int("count", len(ids)).Msg("discovered morpho markets")
	return ids, nil
}
