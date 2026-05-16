package morpho

import (
	"context"

	"github.com/rs/zerolog/log"
	"test-task-edge/internal/types"
)

type marketScanner interface {
	ScanMarkets(ctx context.Context, latestBlock uint64) ([]types.Bytes32, error)
}

type MarketProvider struct {
	scanner marketScanner
	cache   *marketCache
}

func NewMarketProvider(scanner marketScanner) *MarketProvider {
	return &MarketProvider{scanner: scanner, cache: newMarketCache()}
}

func (p *MarketProvider) DiscoverMarkets(ctx context.Context, latestBlock uint64) ([]types.Bytes32, error) {
	if cached, ok := p.cache.get(); ok {
		return cached, nil
	}

	ids, err := p.scanner.ScanMarkets(ctx, latestBlock)
	if err != nil {
		return nil, err
	}

	p.cache.set(ids)

	log.Info().Int("count", len(ids)).Msg("discovered morpho markets")
	return ids, nil
}
