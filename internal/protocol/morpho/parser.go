package morpho

import (
	"context"
	"math/big"
	"sync"

	"github.com/rs/zerolog/log"
	"test-task-edge/internal/protocol"
	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

type tokenGetter interface {
	GetToken(ctx context.Context, token types.Address, block uint64) types.Token
}

type marketDiscoverer interface {
	DiscoverMarkets(ctx context.Context, latestBlock uint64) ([]types.Bytes32, error)
}

type userPositionProvider interface {
	GetUserPosition(ctx context.Context, id types.Bytes32, user types.Address, block uint64) (supplyShares *big.Int, borrowShares *big.Int, collateral *big.Int, err error)
}

type marketConfigProvider interface {
	GetMarketConfig(ctx context.Context, id types.Bytes32, block uint64) (loanToken types.Address, collateralToken types.Address, oracle types.Address, lltv *big.Int, err error)
}

type marketStateProvider interface {
	GetMarketState(ctx context.Context, id types.Bytes32, block uint64) (totalBorrowAssets *big.Int, totalBorrowShares *big.Int, err error)
}

type oraclePriceProvider interface {
	GetOraclePrice(ctx context.Context, oracle types.Address, block uint64) (*big.Int, error)
}

type healthFactorComputer interface {
	GetHealthFactor(collateral, debt, price, lltv *big.Int) float64
}

type Parser struct {
	tokenGetter          tokenGetter
	marketDiscoverer     marketDiscoverer
	userPositionProvider userPositionProvider
	marketConfigProvider marketConfigProvider
	marketStateProvider  marketStateProvider
	oraclePriceProvider  oraclePriceProvider
	hfComputer           healthFactorComputer
	parallelism          int
}

func NewParser(
	tokenGetter tokenGetter,
	marketDiscoverer marketDiscoverer,
	userPositionProvider userPositionProvider,
	marketConfigProvider marketConfigProvider,
	marketStateProvider marketStateProvider,
	oraclePriceProvider oraclePriceProvider,
	hfComputer healthFactorComputer,
	parallelism int,
) *Parser {
	return &Parser{
		tokenGetter:          tokenGetter,
		marketDiscoverer:     marketDiscoverer,
		userPositionProvider: userPositionProvider,
		marketConfigProvider: marketConfigProvider,
		marketStateProvider:  marketStateProvider,
		oraclePriceProvider:  oraclePriceProvider,
		hfComputer:           hfComputer,
		parallelism:          parallelism,
	}
}

func (p *Parser) ParsePositions(ctx context.Context, wallets []types.Address, block types.BlockRef) ([]protocol.Position, error) {
	marketIDs, err := p.marketDiscoverer.DiscoverMarkets(ctx, block.Number)
	if err != nil {
		return nil, err
	}

	log.Info().Int("market_count", len(marketIDs)).Int("wallet_count", len(wallets)).Int("parallelism", p.parallelism).Msg("morpho: checking positions")

	type marketResult struct {
		positions []protocol.Position
	}

	marketCh := make(chan types.Bytes32, len(marketIDs))
	resultCh := make(chan marketResult, len(marketIDs))

	var wg sync.WaitGroup
	for i := 0; i < p.parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for mid := range marketCh {
				positions := p.checkMarket(ctx, mid, wallets, block)
				if len(positions) > 0 {
					resultCh <- marketResult{positions: positions}
				}
			}
		}()
	}

	for _, mid := range marketIDs {
		marketCh <- mid
	}
	close(marketCh)

	wg.Wait()
	close(resultCh)

	var out []protocol.Position
	for r := range resultCh {
		out = append(out, r.positions...)
	}
	return out, nil
}

func (p *Parser) checkMarket(ctx context.Context, marketID types.Bytes32, wallets []types.Address, block types.BlockRef) []protocol.Position {
	log.Debug().Str("market_id", marketID.String()).Msg("morpho: checking market")
	loanToken, collateralToken, oracle, lltv, err := p.marketConfigProvider.GetMarketConfig(ctx, marketID, block.Number)
	if err != nil {
		log.Error().Err(err).Str("market_id", marketID.String()).Msg("failed to read morpho market params")
		return nil
	}
	totalBorrowAssets, totalBorrowShares, err := p.marketStateProvider.GetMarketState(ctx, marketID, block.Number)
	if err != nil {
		log.Error().Err(err).Str("market_id", marketID.String()).Msg("failed to read morpho market state")
		return nil
	}
	loan := p.tokenGetter.GetToken(ctx, loanToken, block.Number)
	collateral := p.tokenGetter.GetToken(ctx, collateralToken, block.Number)

	var out []protocol.Position
	for _, wallet := range wallets {
		_, borrowShares, collat, err := p.userPositionProvider.GetUserPosition(ctx, marketID, wallet, block.Number)
		if err != nil {
			log.Error().Err(err).Str("wallet", wallet.String()).Str("market_id", marketID.String()).Msg("failed to read morpho position")
			continue
		}
		if borrowShares.Sign() == 0 {
			continue
		}

		price, err := p.oraclePriceProvider.GetOraclePrice(ctx, oracle, block.Number)
		if err != nil {
			log.Error().Err(err).Str("market_id", marketID.String()).Str("oracle", oracle.String()).Msg("failed to read morpho oracle price")
			price = big.NewInt(0)
		}

		borrowAssets := utils.SharesToAssets(borrowShares, totalBorrowAssets, totalBorrowShares)
		size := new(big.Int).Add(collat, borrowAssets)
		log.Info().Str("wallet", wallet.String()).Str("market_id", marketID.String()).Float64("position_size", utils.AmountToDecimal(collat, collateral.Decimals)).Msg("morpho: found position")
		out = append(out, protocol.Position{
			Protocol:        "morpho-blue",
			WalletAddress:   wallet.String(),
			MarketID:        marketID.String(),
			CollateralToken: collateral.Symbol,
			DebtToken:       loan.Symbol,
			PositionSize:    utils.AmountToDecimal(size, collateral.Decimals),
			TokenPrice:      utils.AmountToDecimal(price, morphoOraclePriceDecimals),
			HealthFactor:    p.hfComputer.GetHealthFactor(collat, borrowAssets, price, lltv),
			BlockNumber:     block.Number,
			Timestamp:       block.Timestamp,
		})
	}
	return out
}
