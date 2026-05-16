package morpho

import (
	"context"
	"math/big"

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
	GetHealthFactor(collateral, debt, price, lltv *big.Int) string
}

type Parser struct {
	tokenGetter          tokenGetter
	marketDiscoverer     marketDiscoverer
	userPositionProvider userPositionProvider
	marketConfigProvider marketConfigProvider
	marketStateProvider  marketStateProvider
	oraclePriceProvider  oraclePriceProvider
	hfComputer           healthFactorComputer
}

func NewParser(
	tokenGetter tokenGetter,
	marketDiscoverer marketDiscoverer,
	userPositionProvider userPositionProvider,
	marketConfigProvider marketConfigProvider,
	marketStateProvider marketStateProvider,
	oraclePriceProvider oraclePriceProvider,
	hfComputer healthFactorComputer,
) *Parser {
	return &Parser{
		tokenGetter:          tokenGetter,
		marketDiscoverer:     marketDiscoverer,
		userPositionProvider: userPositionProvider,
		marketConfigProvider: marketConfigProvider,
		marketStateProvider:  marketStateProvider,
		oraclePriceProvider:  oraclePriceProvider,
		hfComputer:           hfComputer,
	}
}

func (p *Parser) ParsePositions(ctx context.Context, wallets []types.Address, block types.BlockRef) ([]protocol.Position, error) {
	marketIDs, err := p.marketDiscoverer.DiscoverMarkets(ctx, block.Number)
	if err != nil {
		return nil, err
	}

	var out []protocol.Position
	for _, marketID := range marketIDs {
		loanToken, collateralToken, oracle, lltv, err := p.marketConfigProvider.GetMarketConfig(ctx, marketID, block.Number)
		if err != nil {
			log.Error().Err(err).Str("market_id", marketID.String()).Msg("failed to read morpho market params")
			continue
		}
		totalBorrowAssets, totalBorrowShares, err := p.marketStateProvider.GetMarketState(ctx, marketID, block.Number)
		if err != nil {
			log.Error().Err(err).Str("market_id", marketID.String()).Msg("failed to read morpho market")
			continue
		}
		loan := p.tokenGetter.GetToken(ctx, loanToken, block.Number)
		collateral := p.tokenGetter.GetToken(ctx, collateralToken, block.Number)
		price, err := p.oraclePriceProvider.GetOraclePrice(ctx, oracle, block.Number)
		if err != nil {
			log.Error().Err(err).Str("market_id", marketID.String()).Str("oracle", oracle.String()).Msg("failed to read morpho oracle price")
			price = big.NewInt(0)
		}
		for _, wallet := range wallets {
			supplyShares, borrowShares, collat, err := p.userPositionProvider.GetUserPosition(ctx, marketID, wallet, block.Number)
			if err != nil {
				log.Error().Err(err).Str("wallet", wallet.String()).Str("market_id", marketID.String()).Msg("failed to read morpho position")
				continue
			}
			if collat.Sign() == 0 && borrowShares.Sign() == 0 && supplyShares.Sign() == 0 {
				continue
			}
			borrowAssets := utils.SharesToAssets(borrowShares, totalBorrowAssets, totalBorrowShares)
			size := new(big.Int).Add(collat, borrowAssets)
			out = append(out, protocol.Position{
				Protocol:        "morpho-blue",
				WalletAddress:   wallet.String(),
				MarketID:        marketID.String(),
				CollateralToken: collateral.Symbol,
				DebtToken:       utils.DebtTokenName(loan.Symbol, borrowAssets),
				PositionSize:    utils.AmountToDecimal(size, collateral.Decimals),
				TokenPrice:      utils.AmountToDecimal(price, morphoOraclePriceDecimals),
				HealthFactor:    p.hfComputer.GetHealthFactor(collat, borrowAssets, price, lltv),
				BlockNumber:     block.Number,
				Timestamp:       block.Timestamp,
			})
		}
	}
	return out, nil
}
