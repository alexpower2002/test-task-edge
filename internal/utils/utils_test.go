package utils

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAmountToDecimal(t *testing.T) {
	tests := []struct {
		name     string
		n        *big.Int
		decimals int
		want     float64
	}{
		{name: "six_decimals", n: big.NewInt(123456789), decimals: 6, want: 123.456789},
		{name: "nil_is_zero", n: nil, decimals: 18, want: 0},
		{name: "zero", n: big.NewInt(0), decimals: 2, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, AmountToDecimal(tt.n, tt.decimals), 1e-9)
		})
	}
}

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "lowercase", raw: "0x0000000000000000000000000000000000000001", want: "0x0000000000000000000000000000000000000001"},
		{name: "no_prefix", raw: "0000000000000000000000000000000000000002", want: "0x0000000000000000000000000000000000000002"},
		{name: "trim_space", raw: "  0x0000000000000000000000000000000000000003 ", want: "0x0000000000000000000000000000000000000003"},
		{name: "too_short", raw: "0x01", wantErr: true},
		{name: "bad_hex", raw: "0x000000000000000000000000000000000000000z", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := ParseAddress(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, addr.String())
		})
	}
}

func TestSharesToAssets(t *testing.T) {
	tests := []struct {
		name        string
		shares      string
		totalAssets string
		totalShares string
		want        string
	}{
		{name: "normal", shares: "5", totalAssets: "100", totalShares: "10", want: "50"},
		{name: "zero_shares", shares: "0", totalAssets: "100", totalShares: "10", want: "0"},
		{name: "zero_total_shares", shares: "5", totalAssets: "100", totalShares: "0", want: "0"},
		{name: "nil_shares", shares: "", totalAssets: "1", totalShares: "1", want: "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var shares, totalAssets, totalShares *big.Int
			if tt.shares != "" {
				shares, _ = new(big.Int).SetString(tt.shares, 10)
			}
			totalAssets, _ = new(big.Int).SetString(tt.totalAssets, 10)
			totalShares, _ = new(big.Int).SetString(tt.totalShares, 10)
			assert.Equal(t, tt.want, SharesToAssets(shares, totalAssets, totalShares).String())
		})
	}
}

func TestDebtTokenName(t *testing.T) {
	tests := []struct {
		name   string
		symbol string
		debt   string
		want   string
	}{
		{name: "no_debt", symbol: "WETH", debt: "0", want: ""},
		{name: "with_debt", symbol: "WETH", debt: "1", want: "WETH"},
		{name: "nil_debt", symbol: "USDC", debt: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var debt *big.Int
			if tt.debt != "" {
				debt, _ = new(big.Int).SetString(tt.debt, 10)
			}
			assert.Equal(t, tt.want, DebtTokenName(tt.symbol, debt))
		})
	}
}

func TestParseAddressList(t *testing.T) {
	tests := []struct {
		name    string
		raw     []string
		want    []string
		wantErr bool
	}{
		{
			name: "ok",
			raw:  []string{"0x0000000000000000000000000000000000000001", "0x0000000000000000000000000000000000000002"},
			want: []string{"0x0000000000000000000000000000000000000001", "0x0000000000000000000000000000000000000002"},
		},
		{
			name: "trims_spaces",
			raw:  []string{"  0x0000000000000000000000000000000000000003 ", "0x0000000000000000000000000000000000000004"},
			want: []string{"0x0000000000000000000000000000000000000003", "0x0000000000000000000000000000000000000004"},
		},
		{
			name: "skips_empty_entries",
			raw:  []string{"", "0x0000000000000000000000000000000000000005"},
			want: []string{"0x0000000000000000000000000000000000000005"},
		},
		{
			name:    "error_on_bad_address",
			raw:     []string{"0xbad"},
			wantErr: true,
		},
		{
			name: "empty_input",
			raw:  []string{},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAddressList(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, len(tt.want))
			for i := range got {
				assert.Equal(t, tt.want[i], got[i].String())
			}
		})
	}
}
