package morpho

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

type stubMarketDataSuccess struct {
	result string
}

func (s *stubMarketDataSuccess) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return s.result, nil
}

type stubMarketDataError struct{}

func (s *stubMarketDataError) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return "", errors.New("rpc err")
}

func word(h string) string {
	if len(h) > 64 {
		return h[:64]
	}
	return strings.Repeat("0", 64-len(h)) + h
}

func TestParseMarketState(t *testing.T) {
	assets := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	shares := new(big.Int).Mul(big.NewInt(50), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	assetsWord := word(assets.Text(16))
	sharesWord := word(shares.Text(16))

	makeWords := func(ws ...string) []string {
		return ws
	}

	tests := []struct {
		name          string
		words         []string
		wantAssets    *big.Int
		wantShares    *big.Int
		wantErr       bool
	}{
		{
			name:       "ok",
			words:      makeWords(strings.Repeat("0", 64), strings.Repeat("0", 64), assetsWord, sharesWord),
			wantAssets: assets,
			wantShares: shares,
		},
		{
			name:    "too_few_words",
			words:   makeWords(strings.Repeat("0", 64), strings.Repeat("0", 64), assetsWord),
			wantErr: true,
		},
		{
			name:    "empty",
			words:   nil,
			wantErr: true,
		},
		{
			name:    "invalid_assets",
			words:   makeWords(strings.Repeat("0", 64), strings.Repeat("0", 64), "zz", sharesWord),
			wantErr: true,
		},
		{
			name:    "invalid_shares",
			words:   makeWords(strings.Repeat("0", 64), strings.Repeat("0", 64), assetsWord, "zz"),
			wantErr: true,
		},
		{
			name:       "zero_values",
			words:      makeWords(strings.Repeat("0", 64), strings.Repeat("0", 64), strings.Repeat("0", 64), strings.Repeat("0", 64)),
			wantAssets: new(big.Int),
			wantShares: new(big.Int),
		},
		{
			name:       "more_than_4_words",
			words:      makeWords(strings.Repeat("0", 64), strings.Repeat("0", 64), assetsWord, sharesWord, strings.Repeat("0", 64)),
			wantAssets: assets,
			wantShares: shares,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAssets, gotShares, err := parseMarketState(tt.words)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Zero(t, tt.wantAssets.Cmp(gotAssets))
			assert.Zero(t, tt.wantShares.Cmp(gotShares))
		})
	}
}

func TestGetMarketState(t *testing.T) {
	contractAddr, err := utils.ParseAddress("0x00000000000000000000000000000000000000aa")
	require.NoError(t, err)
	id := types.Bytes32{1}
	assets := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	shares := new(big.Int).Mul(big.NewInt(50), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	makeResponse := func(ws ...string) string {
		return "0x" + strings.Join(ws, "")
	}

	tests := []struct {
		name       string
		caller     ethCaller
		wantAssets *big.Int
		wantShares *big.Int
		wantErr    bool
	}{
		{
			name:       "ok",
			caller:     &stubMarketDataSuccess{result: makeResponse(strings.Repeat("00", 32), strings.Repeat("00", 32), word(assets.Text(16)), word(shares.Text(16)))},
			wantAssets: assets,
			wantShares: shares,
		},
		{
			name:    "eth_call_error",
			caller:  &stubMarketDataError{},
			wantErr: true,
		},
		{
			name:    "decode_words_error",
			caller:  &stubMarketDataSuccess{result: "0xabc"},
			wantErr: true,
		},
		{
			name:    "too_few_words",
			caller:  &stubMarketDataSuccess{result: "0x" + strings.Repeat("00", 32*3)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewMarketDataReader(tt.caller, contractAddr)
			gotAssets, gotShares, err := r.GetMarketState(context.Background(), id, 100)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Zero(t, tt.wantAssets.Cmp(gotAssets))
			assert.Zero(t, tt.wantShares.Cmp(gotShares))
		})
	}
}
