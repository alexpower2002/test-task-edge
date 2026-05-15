package aavev3

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

type stubPriceSuccess struct {
	result string
}

func (s *stubPriceSuccess) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return s.result, nil
}

type stubPriceError struct{}

func (s *stubPriceError) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return "", errors.New("rpc err")
}

func TestGetAssetPrice(t *testing.T) {
	oracle, err := utils.ParseAddress("0x000000000000000000000000000000000000000a")
	require.NoError(t, err)
	asset, err := utils.ParseAddress("0x0000000000000000000000000000000000000001")
	require.NoError(t, err)

	priceWord := "0000000000000000000000000000000000000000000000000000000000000064"

	tests := []struct {
		name     string
		caller   ethCaller
		want     *big.Int
		wantErr  bool
	}{
		{
			name:   "ok",
			caller: &stubPriceSuccess{result: "0x" + priceWord},
			want:   big.NewInt(100),
		},
		{
			name:    "eth_call_error",
			caller:  &stubPriceError{},
			wantErr: true,
		},
		{
			name:    "decode_words_error",
			caller:  &stubPriceSuccess{result: "0xabc"},
			wantErr: true,
		},
		{
			name:    "empty_response",
			caller:  &stubPriceSuccess{result: "0x"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewAssetPricer(tt.caller, oracle)
			got, err := p.GetAssetPrice(context.Background(), asset, 100)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Zero(t, got.Cmp(tt.want))
		})
	}
}
