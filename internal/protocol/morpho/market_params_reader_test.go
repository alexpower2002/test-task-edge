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

type stubMarketParamsSuccess struct {
	result string
}

func (s *stubMarketParamsSuccess) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return s.result, nil
}

type stubMarketParamsError struct{}

func (s *stubMarketParamsError) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return "", errors.New("rpc err")
}

func TestParseMarketConfig(t *testing.T) {
	loanAddr := address(t, "0x00000000000000000000000000000000000000aa")
	collateralAddr := address(t, "0x00000000000000000000000000000000000000bb")
	oracleAddr := address(t, "0x00000000000000000000000000000000000000cc")
	lltvVal := big.NewInt(860000000000000000)

	tests := []struct {
		name          string
		words         []string
		wantLoan      types.Address
		wantCollateral types.Address
		wantOracle    types.Address
		wantLltv      *big.Int
		wantErr       bool
	}{
		{
			name:          "ok",
			words:         []string{types.WordAddress(loanAddr), types.WordAddress(collateralAddr), types.WordAddress(oracleAddr), strings.Repeat("0", 64), word(lltvVal.Text(16))},
			wantLoan:      loanAddr,
			wantCollateral: collateralAddr,
			wantOracle:    oracleAddr,
			wantLltv:      lltvVal,
		},
		{
			name:    "too_few_words",
			words:   []string{types.WordAddress(loanAddr), types.WordAddress(collateralAddr), types.WordAddress(oracleAddr)},
			wantErr: true,
		},
		{
			name:    "empty",
			words:   nil,
			wantErr: true,
		},
		{
			name:    "invalid_loan",
			words:   []string{"zz", types.WordAddress(collateralAddr), types.WordAddress(oracleAddr), strings.Repeat("0", 64), word(lltvVal.Text(16))},
			wantErr: true,
		},
		{
			name:    "invalid_collateral",
			words:   []string{types.WordAddress(loanAddr), "zz", types.WordAddress(oracleAddr), strings.Repeat("0", 64), word(lltvVal.Text(16))},
			wantErr: true,
		},
		{
			name:    "invalid_oracle",
			words:   []string{types.WordAddress(loanAddr), types.WordAddress(collateralAddr), "zz", strings.Repeat("0", 64), word(lltvVal.Text(16))},
			wantErr: true,
		},
		{
			name:    "invalid_lltv",
			words:   []string{types.WordAddress(loanAddr), types.WordAddress(collateralAddr), types.WordAddress(oracleAddr), strings.Repeat("0", 64), "zz"},
			wantErr: true,
		},
		{
			name:          "more_than_5_words",
			words:         []string{types.WordAddress(loanAddr), types.WordAddress(collateralAddr), types.WordAddress(oracleAddr), strings.Repeat("0", 64), word(lltvVal.Text(16)), strings.Repeat("0", 64)},
			wantLoan:      loanAddr,
			wantCollateral: collateralAddr,
			wantOracle:    oracleAddr,
			wantLltv:      lltvVal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLoan, gotCollateral, gotOracle, gotLltv, err := parseMarketConfig(tt.words)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantLoan, gotLoan)
			assert.Equal(t, tt.wantCollateral, gotCollateral)
			assert.Equal(t, tt.wantOracle, gotOracle)
			assert.Zero(t, tt.wantLltv.Cmp(gotLltv))
		})
	}
}

func TestGetMarketConfig(t *testing.T) {
	contractAddr, err := utils.ParseAddress("0x00000000000000000000000000000000000000dd")
	require.NoError(t, err)
	id := types.Bytes32{1}

	loanAddr := address(t, "0x00000000000000000000000000000000000000aa")
	collateralAddr := address(t, "0x00000000000000000000000000000000000000bb")
	oracleAddr := address(t, "0x00000000000000000000000000000000000000cc")
	lltvVal := big.NewInt(860000000000000000)

	makeResponse := func(ws ...string) string {
		return "0x" + strings.Join(ws, "")
	}

	tests := []struct {
		name          string
		caller        ethCaller
		wantLoan      types.Address
		wantCollateral types.Address
		wantOracle    types.Address
		wantLltv      *big.Int
		wantErr       bool
	}{
		{
			name:          "ok",
			caller:        &stubMarketParamsSuccess{result: makeResponse(types.WordAddress(loanAddr), types.WordAddress(collateralAddr), types.WordAddress(oracleAddr), strings.Repeat("00", 32), word(lltvVal.Text(16)))},
			wantLoan:      loanAddr,
			wantCollateral: collateralAddr,
			wantOracle:    oracleAddr,
			wantLltv:      lltvVal,
		},
		{
			name:    "eth_call_error",
			caller:  &stubMarketParamsError{},
			wantErr: true,
		},
		{
			name:    "decode_words_error",
			caller:  &stubMarketParamsSuccess{result: "0xabc"},
			wantErr: true,
		},
		{
			name:    "too_few_words",
			caller:  &stubMarketParamsSuccess{result: "0x" + strings.Repeat("00", 32*3)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewMarketParamsReader(tt.caller, contractAddr)
			gotLoan, gotCollateral, gotOracle, gotLltv, err := r.GetMarketConfig(context.Background(), id, 100)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantLoan, gotLoan)
			assert.Equal(t, tt.wantCollateral, gotCollateral)
			assert.Equal(t, tt.wantOracle, gotOracle)
			assert.Zero(t, tt.wantLltv.Cmp(gotLltv))
		})
	}
}
