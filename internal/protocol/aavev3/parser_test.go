package aavev3

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

type stubReservesOK struct {
	reserves []types.Address
}

func (s *stubReservesOK) DiscoverReserves(_ context.Context, _ uint64) ([]types.Address, error) {
	return s.reserves, nil
}

type stubReservesError struct{}

func (s *stubReservesError) DiscoverReserves(_ context.Context, _ uint64) ([]types.Address, error) {
	return nil, errors.New("discover err")
}

type stubHealthOK struct {
	hf float64
}

func (s *stubHealthOK) GetHealthFactor(_ context.Context, _ types.Address, _ uint64) (float64, error) {
	return s.hf, nil
}

type stubHFError struct{}

func (s *stubHFError) GetHealthFactor(_ context.Context, _ types.Address, _ uint64) (float64, error) {
	return 0, errors.New("health err")
}

type stubUserStateOK struct {
	deposit *big.Int
	borrow  *big.Int
}

func (s *stubUserStateOK) GetUserState(_ context.Context, _, _ types.Address, _ uint64) (*big.Int, *big.Int, error) {
	return s.deposit, s.borrow, nil
}

type stubUserStateError struct{}

func (s *stubUserStateError) GetUserState(_ context.Context, _, _ types.Address, _ uint64) (*big.Int, *big.Int, error) {
	return nil, nil, errors.New("user state err")
}

type stubPriceOK struct {
	price *big.Int
}

func (s *stubPriceOK) GetAssetPrice(_ context.Context, _ types.Address, _ uint64) (*big.Int, error) {
	return s.price, nil
}

type stubAssetPriceError struct{}

func (s *stubAssetPriceError) GetAssetPrice(_ context.Context, _ types.Address, _ uint64) (*big.Int, error) {
	return nil, errors.New("price err")
}

type stubTokenOK struct {
	token types.Token
}

func (s *stubTokenOK) GetToken(_ context.Context, _ types.Address, _ uint64) types.Token {
	return s.token
}

func TestParsePositions(t *testing.T) {
	pool := address(t, "0x00000000000000000000000000000000000000bb")
	wallet := address(t, "0x0000000000000000000000000000000000000001")
	reserve := address(t, "0x00000000000000000000000000000000000000aa")
	reserve2 := address(t, "0x00000000000000000000000000000000000000cc")

	deposit := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	price := new(big.Int).Mul(big.NewInt(200), new(big.Int).Exp(big.NewInt(10), big.NewInt(8), nil))

	token := types.Token{Address: reserve, Symbol: "USDC", Decimals: 18}

	block := types.BlockRef{Number: 42, Timestamp: 1000}

	tests := []struct {
		name      string
		reserves  reserveDiscoverer
		health    healthFactorProvider
		userState userStateProvider
		price     assetPriceProvider
		token     tokenGetter
		wallets   []types.Address
		wantLen   int
		wantErr   bool
	}{
		{
			name:      "single_wallet_single_reserve",
			reserves:  &stubReservesOK{reserves: []types.Address{reserve}},
			health:    &stubHealthOK{hf: 1.5},
			userState: &stubUserStateOK{deposit: deposit, borrow: deposit},
			price:     &stubPriceOK{price: price},
			token:     &stubTokenOK{token: token},
			wallets:   []types.Address{wallet},
			wantLen:   1,
		},
		{
			name:      "multiple_wallets",
			reserves:  &stubReservesOK{reserves: []types.Address{reserve}},
			health:    &stubHealthOK{hf: 1.2},
			userState: &stubUserStateOK{deposit: deposit, borrow: deposit},
			price:     &stubPriceOK{price: price},
			token:     &stubTokenOK{token: token},
			wallets:   []types.Address{wallet, address(t, "0x0000000000000000000000000000000000000002")},
			wantLen:   2,
		},
		{
			name:      "discover_error",
			reserves:  &stubReservesError{},
			wantErr:   true,
		},
		{
			name:      "health_factor_error",
			reserves:  &stubReservesOK{reserves: []types.Address{reserve}},
			health:    &stubHFError{},
			userState: &stubUserStateOK{deposit: deposit, borrow: deposit},
			price:     &stubPriceOK{price: price},
			token:     &stubTokenOK{token: token},
			wallets:   []types.Address{wallet},
			wantLen:   1,
		},
		{
			name:      "user_state_error",
			reserves:  &stubReservesOK{reserves: []types.Address{reserve}},
			health:    &stubHealthOK{hf: 1.5},
			userState: &stubUserStateError{},
			price:     &stubPriceOK{price: price},
			token:     &stubTokenOK{token: token},
			wallets:   []types.Address{wallet},
			wantLen:   0,
		},
		{
			name:      "both_zero",
			reserves:  &stubReservesOK{reserves: []types.Address{reserve}},
			health:    &stubHealthOK{hf: 1.5},
			userState: &stubUserStateOK{deposit: big.NewInt(0), borrow: big.NewInt(0)},
			price:     &stubPriceOK{price: price},
			token:     &stubTokenOK{token: token},
			wallets:   []types.Address{wallet},
			wantLen:   0,
		},
		{
			name:      "price_error",
			reserves:  &stubReservesOK{reserves: []types.Address{reserve}},
			health:    &stubHealthOK{hf: 1.5},
			userState: &stubUserStateOK{deposit: deposit, borrow: deposit},
			price:     &stubAssetPriceError{},
			token:     &stubTokenOK{token: token},
			wallets:   []types.Address{wallet},
			wantLen:   1,
		},
		{
			name:      "multiple_reserves",
			reserves:  &stubReservesOK{reserves: []types.Address{reserve, reserve2}},
			health:    &stubHealthOK{hf: 1.5},
			userState: &stubUserStateOK{deposit: deposit, borrow: deposit},
			price:     &stubPriceOK{price: price},
			token:     &stubTokenOK{token: token},
			wallets:   []types.Address{wallet},
			wantLen:   2,
		},
		{
			name:      "borrow_present",
			reserves:  &stubReservesOK{reserves: []types.Address{reserve}},
			health:    &stubHealthOK{hf: 0.8},
			userState: &stubUserStateOK{deposit: big.NewInt(0), borrow: deposit},
			price:     &stubPriceOK{price: price},
			token:     &stubTokenOK{token: token},
			wallets:   []types.Address{wallet},
			wantLen:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(pool, tt.token, tt.reserves, tt.userState, tt.price, tt.health)
			got, err := p.ParsePositions(context.Background(), tt.wallets, block)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, tt.wantLen)
			if tt.wantLen > 0 {
				for _, pos := range got {
					assert.Equal(t, "aave-v3", pos.Protocol)
					assert.Equal(t, block.Number, pos.BlockNumber)
					assert.Equal(t, block.Timestamp, pos.Timestamp)
					assert.True(t, strings.HasPrefix(pos.MarketID, pool.String()+":"))
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
