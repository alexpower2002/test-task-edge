package morpho

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

type stubMarketDiscoverSuccess struct {
	ids []types.Bytes32
}

func (s *stubMarketDiscoverSuccess) DiscoverMarkets(_ context.Context, _ uint64) ([]types.Bytes32, error) {
	return s.ids, nil
}

type stubMarketDiscoverError struct{}

func (s *stubMarketDiscoverError) DiscoverMarkets(_ context.Context, _ uint64) ([]types.Bytes32, error) {
	return nil, errors.New("discover err")
}

type stubMarketConfigSuccess struct {
	loanToken       types.Address
	collateralToken types.Address
	oracle          types.Address
	lltv            *big.Int
}

func (s *stubMarketConfigSuccess) GetMarketConfig(_ context.Context, _ types.Bytes32, _ uint64) (types.Address, types.Address, types.Address, *big.Int, error) {
	return s.loanToken, s.collateralToken, s.oracle, s.lltv, nil
}

type stubMarketConfigError struct{}

func (s *stubMarketConfigError) GetMarketConfig(_ context.Context, _ types.Bytes32, _ uint64) (types.Address, types.Address, types.Address, *big.Int, error) {
	return types.Address{}, types.Address{}, types.Address{}, nil, errors.New("config err")
}

type stubMarketStateSuccess struct {
	totalBorrowAssets *big.Int
	totalBorrowShares *big.Int
}

func (s *stubMarketStateSuccess) GetMarketState(_ context.Context, _ types.Bytes32, _ uint64) (*big.Int, *big.Int, error) {
	return s.totalBorrowAssets, s.totalBorrowShares, nil
}

type stubMarketStateError struct{}

func (s *stubMarketStateError) GetMarketState(_ context.Context, _ types.Bytes32, _ uint64) (*big.Int, *big.Int, error) {
	return nil, nil, errors.New("state err")
}

type stubOraclePriceSuccess struct {
	price *big.Int
}

func (s *stubOraclePriceSuccess) GetOraclePrice(_ context.Context, _ types.Address, _ uint64) (*big.Int, error) {
	return s.price, nil
}

type stubOraclePriceError struct{}

func (s *stubOraclePriceError) GetOraclePrice(_ context.Context, _ types.Address, _ uint64) (*big.Int, error) {
	return nil, errors.New("price err")
}

type stubPositionSuccess struct {
	supplyShares *big.Int
	borrowShares *big.Int
	collateral   *big.Int
}

func (s *stubPositionSuccess) GetUserPosition(_ context.Context, _ types.Bytes32, _ types.Address, _ uint64) (*big.Int, *big.Int, *big.Int, error) {
	return s.supplyShares, s.borrowShares, s.collateral, nil
}

type stubPositionError struct{}

func (s *stubPositionError) GetUserPosition(_ context.Context, _ types.Bytes32, _ types.Address, _ uint64) (*big.Int, *big.Int, *big.Int, error) {
	return nil, nil, nil, errors.New("position err")
}

type stubTokenGetterSuccess struct {
	token types.Token
}

func (s *stubTokenGetterSuccess) GetToken(_ context.Context, _ types.Address, _ uint64) types.Token {
	return s.token
}

func TestParsePositions(t *testing.T) {
	wallet1 := address(t, "0x0000000000000000000000000000000000000001")
	wallet2 := address(t, "0x0000000000000000000000000000000000000002")
	loanTokenAddr := address(t, "0x00000000000000000000000000000000000000aa")
	collateralTokenAddr := address(t, "0x00000000000000000000000000000000000000bb")
	oracleAddr := address(t, "0x00000000000000000000000000000000000000cc")

	id1 := types.Bytes32{1}
	id2 := types.Bytes32{2}

	pow18 := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	pow36 := new(big.Int).Exp(big.NewInt(10), big.NewInt(36), nil)
	lltv := new(big.Int).SetInt64(860000000000000000)

	collateralAmount := new(big.Int).Mul(big.NewInt(5), pow18)
	price := pow36

	loanToken := types.Token{Address: loanTokenAddr, Symbol: "WETH", Decimals: 18}

	block := types.BlockRef{Number: 42, Timestamp: 1000}

	tests := []struct {
		name       string
		discoverer marketDiscoverer
		config     marketConfigProvider
		state      marketStateProvider
		pricer     oraclePriceProvider
		position   userPositionProvider
		token      tokenGetter
		wallets    []types.Address
		wantLen    int
		wantErr    bool
	}{
		{
			name:       "single_market_single_wallet",
			discoverer: &stubMarketDiscoverSuccess{ids: []types.Bytes32{id1}},
			config:     &stubMarketConfigSuccess{loanToken: loanTokenAddr, collateralToken: collateralTokenAddr, oracle: oracleAddr, lltv: lltv},
			state:      &stubMarketStateSuccess{totalBorrowAssets: pow18, totalBorrowShares: pow18},
			pricer:     &stubOraclePriceSuccess{price: price},
			position:   &stubPositionSuccess{supplyShares: pow18, borrowShares: pow18, collateral: collateralAmount},
			token:      &stubTokenGetterSuccess{token: loanToken},
			wallets:    []types.Address{wallet1},
			wantLen:    1,
		},
		{
			name:       "multiple_wallets",
			discoverer: &stubMarketDiscoverSuccess{ids: []types.Bytes32{id1}},
			config:     &stubMarketConfigSuccess{loanToken: loanTokenAddr, collateralToken: collateralTokenAddr, oracle: oracleAddr, lltv: lltv},
			state:      &stubMarketStateSuccess{totalBorrowAssets: pow18, totalBorrowShares: pow18},
			pricer:     &stubOraclePriceSuccess{price: price},
			position:   &stubPositionSuccess{supplyShares: pow18, borrowShares: pow18, collateral: collateralAmount},
			token:      &stubTokenGetterSuccess{token: loanToken},
			wallets:    []types.Address{wallet1, wallet2},
			wantLen:    2,
		},
		{
			name:       "discover_error",
			discoverer: &stubMarketDiscoverError{},
			wantErr:    true,
		},
		{
			name:       "market_config_error",
			discoverer: &stubMarketDiscoverSuccess{ids: []types.Bytes32{id1}},
			config:     &stubMarketConfigError{},
			wantLen:    0,
		},
		{
			name:       "market_state_error",
			discoverer: &stubMarketDiscoverSuccess{ids: []types.Bytes32{id1}},
			config:     &stubMarketConfigSuccess{loanToken: loanTokenAddr, collateralToken: collateralTokenAddr, oracle: oracleAddr, lltv: lltv},
			state:      &stubMarketStateError{},
			wantLen:    0,
		},
		{
			name:       "oracle_price_error",
			discoverer: &stubMarketDiscoverSuccess{ids: []types.Bytes32{id1}},
			config:     &stubMarketConfigSuccess{loanToken: loanTokenAddr, collateralToken: collateralTokenAddr, oracle: oracleAddr, lltv: lltv},
			state:      &stubMarketStateSuccess{totalBorrowAssets: pow18, totalBorrowShares: pow18},
			pricer:     &stubOraclePriceError{},
			position:   &stubPositionSuccess{supplyShares: pow18, borrowShares: pow18, collateral: collateralAmount},
			token:      &stubTokenGetterSuccess{token: loanToken},
			wallets:    []types.Address{wallet1},
			wantLen:    1,
		},
		{
			name:       "position_error",
			discoverer: &stubMarketDiscoverSuccess{ids: []types.Bytes32{id1}},
			config:     &stubMarketConfigSuccess{loanToken: loanTokenAddr, collateralToken: collateralTokenAddr, oracle: oracleAddr, lltv: lltv},
			state:      &stubMarketStateSuccess{totalBorrowAssets: pow18, totalBorrowShares: pow18},
			pricer:     &stubOraclePriceSuccess{price: price},
			position:   &stubPositionError{},
			token:      &stubTokenGetterSuccess{token: loanToken},
			wallets:    []types.Address{wallet1},
			wantLen:    0,
		},
		{
			name:       "position_all_zero",
			discoverer: &stubMarketDiscoverSuccess{ids: []types.Bytes32{id1}},
			config:     &stubMarketConfigSuccess{loanToken: loanTokenAddr, collateralToken: collateralTokenAddr, oracle: oracleAddr, lltv: lltv},
			state:      &stubMarketStateSuccess{totalBorrowAssets: pow18, totalBorrowShares: pow18},
			pricer:     &stubOraclePriceSuccess{price: price},
			position:   &stubPositionSuccess{supplyShares: big.NewInt(0), borrowShares: big.NewInt(0), collateral: big.NewInt(0)},
			token:      &stubTokenGetterSuccess{token: loanToken},
			wallets:    []types.Address{wallet1},
			wantLen:    0,
		},
		{
			name:       "multiple_markets",
			discoverer: &stubMarketDiscoverSuccess{ids: []types.Bytes32{id1, id2}},
			config:     &stubMarketConfigSuccess{loanToken: loanTokenAddr, collateralToken: collateralTokenAddr, oracle: oracleAddr, lltv: lltv},
			state:      &stubMarketStateSuccess{totalBorrowAssets: pow18, totalBorrowShares: pow18},
			pricer:     &stubOraclePriceSuccess{price: price},
			position:   &stubPositionSuccess{supplyShares: pow18, borrowShares: pow18, collateral: collateralAmount},
			token:      &stubTokenGetterSuccess{token: loanToken},
			wallets:    []types.Address{wallet1},
			wantLen:    2,
		},
		{
			name:       "borrow_present",
			discoverer: &stubMarketDiscoverSuccess{ids: []types.Bytes32{id1}},
			config:     &stubMarketConfigSuccess{loanToken: loanTokenAddr, collateralToken: collateralTokenAddr, oracle: oracleAddr, lltv: lltv},
			state:      &stubMarketStateSuccess{totalBorrowAssets: pow18, totalBorrowShares: pow18},
			pricer:     &stubOraclePriceSuccess{price: price},
			position:   &stubPositionSuccess{supplyShares: big.NewInt(0), borrowShares: pow18, collateral: big.NewInt(0)},
			token:      &stubTokenGetterSuccess{token: loanToken},
			wallets:    []types.Address{wallet1},
			wantLen:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.token, tt.discoverer, tt.position, tt.config, tt.state, tt.pricer, NewHealthFactorReader())
			got, err := p.ParsePositions(context.Background(), tt.wallets, block)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, tt.wantLen)
			if tt.wantLen > 0 {
				for _, pos := range got {
					assert.Equal(t, "morpho-blue", pos.Protocol)
					assert.Equal(t, block.Number, pos.BlockNumber)
					assert.Equal(t, block.Timestamp, pos.Timestamp)
				}
			}
		})
	}
}

func address(t *testing.T, s string) types.Address {
	t.Helper()
	a, err := utils.ParseAddress(s)
	require.NoError(t, err)
	return a
}
