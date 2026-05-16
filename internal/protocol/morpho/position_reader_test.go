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

type stubPositionReaderSuccess struct {
	result string
}

func (s *stubPositionReaderSuccess) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return s.result, nil
}

type stubPositionReaderError struct{}

func (s *stubPositionReaderError) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return "", errors.New("rpc err")
}

func TestParsePositionData(t *testing.T) {
	supply := new(big.Int).Mul(big.NewInt(10), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	borrow := new(big.Int).Mul(big.NewInt(5), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	collat := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	tests := []struct {
		name         string
		words        []string
		wantSupply   *big.Int
		wantBorrow   *big.Int
		wantCollat   *big.Int
		wantErr      bool
	}{
		{
			name:       "ok",
			words:      []string{word(supply.Text(16)), word(borrow.Text(16)), word(collat.Text(16))},
			wantSupply: supply,
			wantBorrow: borrow,
			wantCollat: collat,
		},
		{
			name:    "too_few_words",
			words:   []string{word(supply.Text(16)), word(borrow.Text(16))},
			wantErr: true,
		},
		{
			name:    "empty",
			words:   nil,
			wantErr: true,
		},
		{
			name:    "invalid_supply",
			words:   []string{"zz", word(borrow.Text(16)), word(collat.Text(16))},
			wantErr: true,
		},
		{
			name:    "invalid_borrow",
			words:   []string{word(supply.Text(16)), "zz", word(collat.Text(16))},
			wantErr: true,
		},
		{
			name:    "invalid_collateral",
			words:   []string{word(supply.Text(16)), word(borrow.Text(16)), "zz"},
			wantErr: true,
		},
		{
			name:       "zero_values",
			words:      []string{strings.Repeat("0", 64), strings.Repeat("0", 64), strings.Repeat("0", 64)},
			wantSupply: new(big.Int),
			wantBorrow: new(big.Int),
			wantCollat: new(big.Int),
		},
		{
			name:       "more_than_3_words",
			words:      []string{word(supply.Text(16)), word(borrow.Text(16)), word(collat.Text(16)), strings.Repeat("0", 64)},
			wantSupply: supply,
			wantBorrow: borrow,
			wantCollat: collat,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSupply, gotBorrow, gotCollat, err := parsePositionData(tt.words)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Zero(t, tt.wantSupply.Cmp(gotSupply))
			assert.Zero(t, tt.wantBorrow.Cmp(gotBorrow))
			assert.Zero(t, tt.wantCollat.Cmp(gotCollat))
		})
	}
}

func TestGetUserPosition(t *testing.T) {
	contractAddr, err := utils.ParseAddress("0x00000000000000000000000000000000000000aa")
	require.NoError(t, err)
	id := types.Bytes32{1}
	userAddr, err := utils.ParseAddress("0x0000000000000000000000000000000000000001")
	require.NoError(t, err)

	supply := new(big.Int).Mul(big.NewInt(10), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	borrow := new(big.Int).Mul(big.NewInt(5), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	collat := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	makeResponse := func(ws ...string) string {
		return "0x" + strings.Join(ws, "")
	}

	tests := []struct {
		name       string
		caller     ethCaller
		wantSupply *big.Int
		wantBorrow *big.Int
		wantCollat *big.Int
		wantErr    bool
	}{
		{
			name:       "ok",
			caller:     &stubPositionReaderSuccess{result: makeResponse(word(supply.Text(16)), word(borrow.Text(16)), word(collat.Text(16)))},
			wantSupply: supply,
			wantBorrow: borrow,
			wantCollat: collat,
		},
		{
			name:    "eth_call_error",
			caller:  &stubPositionReaderError{},
			wantErr: true,
		},
		{
			name:    "decode_words_error",
			caller:  &stubPositionReaderSuccess{result: "0xabc"},
			wantErr: true,
		},
		{
			name:    "too_few_words",
			caller:  &stubPositionReaderSuccess{result: "0x" + strings.Repeat("00", 32*2)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewPositionReader(tt.caller, contractAddr)
			gotSupply, gotBorrow, gotCollat, err := r.GetUserPosition(context.Background(), id, userAddr, 100)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Zero(t, tt.wantSupply.Cmp(gotSupply))
			assert.Zero(t, tt.wantBorrow.Cmp(gotBorrow))
			assert.Zero(t, tt.wantCollat.Cmp(gotCollat))
		})
	}
}
