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

type stubOraclePricerSuccess struct {
	result string
}

func (s *stubOraclePricerSuccess) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return s.result, nil
}

type stubOraclePricerError struct{}

func (s *stubOraclePricerError) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return "", errors.New("rpc err")
}

func TestParseOraclePrice(t *testing.T) {
	priceVal := new(big.Int).Mul(big.NewInt(200), new(big.Int).Exp(big.NewInt(10), big.NewInt(8), nil))

	tests := []struct {
		name    string
		words   []string
		want    *big.Int
		wantErr bool
	}{
		{
			name:  "ok",
			words: []string{word(priceVal.Text(16))},
			want:  priceVal,
		},
		{
			name:    "empty",
			words:   nil,
			wantErr: true,
		},
		{
			name:    "invalid_word",
			words:   []string{"zz"},
			wantErr: true,
		},
		{
			name:  "more_than_1_word",
			words: []string{word(priceVal.Text(16)), strings.Repeat("0", 64)},
			want:  priceVal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOraclePrice(tt.words)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Zero(t, tt.want.Cmp(got))
		})
	}
}

func TestGetOraclePrice(t *testing.T) {
	oracleAddr, err := utils.ParseAddress("0x00000000000000000000000000000000000000aa")
	require.NoError(t, err)
	priceVal := new(big.Int).Mul(big.NewInt(200), new(big.Int).Exp(big.NewInt(10), big.NewInt(8), nil))

	tests := []struct {
		name    string
		caller  ethCaller
		oracle  types.Address
		want    *big.Int
		wantErr bool
	}{
		{
			name:   "zero_oracle",
			caller: &stubOraclePricerSuccess{},
			oracle: types.Address{},
			want:   big.NewInt(0),
		},
		{
			name:   "ok",
			caller: &stubOraclePricerSuccess{result: "0x" + word(priceVal.Text(16))},
			oracle: oracleAddr,
			want:   priceVal,
		},
		{
			name:    "eth_call_error",
			caller:  &stubOraclePricerError{},
			oracle:  oracleAddr,
			wantErr: true,
		},
		{
			name:    "decode_words_error",
			caller:  &stubOraclePricerSuccess{result: "0xabc"},
			oracle:  oracleAddr,
			wantErr: true,
		},
		{
			name:    "empty_response",
			caller:  &stubOraclePricerSuccess{result: "0x"},
			oracle:  oracleAddr,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewOraclePricer(tt.caller)
			got, err := r.GetOraclePrice(context.Background(), tt.oracle, 100)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Zero(t, tt.want.Cmp(got))
		})
	}
}
