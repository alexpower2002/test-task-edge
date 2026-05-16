package ethereum

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

type stubTokenResults struct {
	results map[string]string
}

func (s *stubTokenResults) EthCall(_ context.Context, _ types.Address, data string, _ uint64) (string, error) {
	return s.results[data], nil
}

type stubTokenError struct{}

func (s *stubTokenError) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return "", errors.New("rpc err")
}

func TestReadToken(t *testing.T) {
	addr := address(t, "0x0000000000000000000000000000000000000001")
	decimals6 := "0x0000000000000000000000000000000000000000000000000000000000000006"
	symbolUSDC := "0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000045553444300000000000000000000000000000000000000000000000000000000"

	tests := []struct {
		name   string
		caller ethCaller
		want   types.Token
	}{
		{
			name: "all_ok",
			caller: &stubTokenResults{results: map[string]string{
				"0x313ce567": decimals6,
				"0x95d89b41": symbolUSDC,
			}},
			want: types.Token{Address: addr, Symbol: "USDC", Decimals: 6},
		},
		{
			name:   "both_fail",
			caller: &stubTokenError{},
			want:   types.Token{Address: addr, Symbol: addr.String(), Decimals: 18},
		},
		{
			name: "decimals_fail_symbol_ok",
			caller: &stubTokenResults{results: map[string]string{
				"0x95d89b41": symbolUSDC,
			}},
			want: types.Token{Address: addr, Symbol: "USDC", Decimals: 18},
		},
		{
			name: "symbol_fail_decimals_ok",
			caller: &stubTokenResults{results: map[string]string{
				"0x313ce567": decimals6,
			}},
			want: types.Token{Address: addr, Symbol: addr.String(), Decimals: 6},
		},
		{
			name: "decimals_bad_word",
			caller: &stubTokenResults{results: map[string]string{
				"0x313ce567": "0xdeadbeef",
				"0x95d89b41": symbolUSDC,
			}},
			want: types.Token{Address: addr, Symbol: "USDC", Decimals: 18},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTokenReader(tt.caller)
			got := r.ReadToken(context.Background(), addr, 100)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadToken_staticString(t *testing.T) {
	addr := address(t, "0x0000000000000000000000000000000000000002")
	decimals8 := "0x0000000000000000000000000000000000000000000000000000000000000008"

	r := NewTokenReader(&stubTokenResults{results: map[string]string{
		"0x313ce567": decimals8,
		"0x95d89b41": "0x55534443",
	}})
	got := r.ReadToken(context.Background(), addr, 100)
	assert.Equal(t, 8, got.Decimals)
	assert.Equal(t, "USDC", got.Symbol)
}

func address(t *testing.T, s string) types.Address {
	t.Helper()
	a, err := utils.ParseAddress(s)
	require.NoError(t, err)
	return a
}
