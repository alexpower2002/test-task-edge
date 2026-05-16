package aavev3

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

type reserveDiscoverer interface {
	DiscoverReserves(ctx context.Context, block uint64) ([]types.Address, error)
}

type userStateProvider interface {
	GetUserState(ctx context.Context, asset, user types.Address, block uint64) (deposit *big.Int, borrow *big.Int, err error)
}

type assetPriceProvider interface {
	GetAssetPrice(ctx context.Context, asset types.Address, block uint64) (*big.Int, error)
}

type healthFactorProvider interface {
	GetHealthFactor(ctx context.Context, user types.Address, block uint64) (float64, error)
}

type Parser struct {
	pool                 types.Address
	tokenGetter          tokenGetter
	reserveDiscoverer    reserveDiscoverer
	userStateProvider    userStateProvider
	assetPriceProvider   assetPriceProvider
	healthFactorProvider healthFactorProvider
}

func NewParser(pool types.Address, tokenGetter tokenGetter, reserveDiscoverer reserveDiscoverer, userStateProvider userStateProvider, assetPriceProvider assetPriceProvider, healthFactorProvider healthFactorProvider) *Parser {
	return &Parser{
		pool:                 pool,
		tokenGetter:          tokenGetter,
		reserveDiscoverer:    reserveDiscoverer,
		userStateProvider:    userStateProvider,
		assetPriceProvider:   assetPriceProvider,
		healthFactorProvider: healthFactorProvider,
	}
}

func (p *Parser) ParsePositions(ctx context.Context, wallets []types.Address, block types.BlockRef) ([]protocol.Position, error) {
	reserves, err := p.reserveDiscoverer.DiscoverReserves(ctx, block.Number)
	if err != nil {
		return nil, err
	}

	var out []protocol.Position
	for _, wallet := range wallets {
		healthFactor, err := p.healthFactorProvider.GetHealthFactor(ctx, wallet, block.Number)
		if err != nil {
			log.Error().Err(err).Str("wallet", wallet.String()).Msg("failed to read aave health factor")
			healthFactor = 0
		}
		for _, reserve := range reserves {
			aToken, variableDebt, err := p.userStateProvider.GetUserState(ctx, reserve, wallet, block.Number)
			if err != nil {
				log.Error().Err(err).Str("wallet", wallet.String()).Str("reserve", reserve.String()).Msg("failed to read aave reserve data")
				continue
			}
			if variableDebt.Sign() == 0 {
				continue
			}
			price, err := p.assetPriceProvider.GetAssetPrice(ctx, reserve, block.Number)
			if err != nil {
				log.Error().Err(err).Str("reserve", reserve.String()).Msg("failed to read aave asset price")
				price = big.NewInt(0)
			}
			token := p.tokenGetter.GetToken(ctx, reserve, block.Number)
			size := new(big.Int).Add(aToken, variableDebt)
			out = append(out, protocol.Position{
				Protocol:        "aave-v3",
				WalletAddress:   wallet.String(),
				MarketID:        p.pool.String() + ":" + reserve.String(),
				CollateralToken: token.Symbol,
				DebtToken:       token.Symbol,
				PositionSize:    utils.AmountToDecimal(size, token.Decimals),
				TokenPrice:      utils.AmountToDecimal(price, 8),
				HealthFactor:    healthFactor,
				BlockNumber:     block.Number,
				Timestamp:       block.Timestamp,
			})
		}
	}
	return out, nil
}
