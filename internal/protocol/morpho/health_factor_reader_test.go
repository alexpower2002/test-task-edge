package morpho

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeHealthFactor(t *testing.T) {
	pow18 := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	pow36 := new(big.Int).Exp(big.NewInt(10), big.NewInt(36), nil)
	lltv := big.NewInt(860000000000000000)

	tests := []struct {
		name       string
		collateral *big.Int
		debt       *big.Int
		price      *big.Int
		lltv       *big.Int
		want       float64
	}{
		{
			name:       "numeric",
			collateral: pow18,
			debt:       pow18,
			price:      pow36,
			lltv:       lltv,
			want:       0.86,
		},
		{
			name:       "zero_debt",
			collateral: pow18,
			debt:       big.NewInt(0),
			price:      pow36,
			lltv:       lltv,
			want:       0,
		},
		{
			name:       "nil_debt",
			collateral: pow18,
			debt:       nil,
			price:      pow36,
			lltv:       lltv,
			want:       0,
		},
		{
			name:       "nil_collateral",
			collateral: nil,
			debt:       pow18,
			price:      pow36,
			lltv:       lltv,
			want:       0,
		},
		{
			name:       "nil_price",
			collateral: pow18,
			debt:       pow18,
			price:      nil,
			lltv:       lltv,
			want:       0,
		},
		{
			name:       "nil_lltv",
			collateral: pow18,
			debt:       pow18,
			price:      pow36,
			lltv:       nil,
			want:       0,
		},
		{
			name:       "zero_collateral",
			collateral: big.NewInt(0),
			debt:       pow18,
			price:      pow36,
			lltv:       lltv,
			want:       0,
		},
		{
			name:       "zero_price",
			collateral: pow18,
			debt:       pow18,
			price:      big.NewInt(0),
			lltv:       lltv,
			want:       0,
		},
		{
			name:       "exact_ratio",
			collateral: new(big.Int).Mul(big.NewInt(100), pow18),
			debt:       new(big.Int).Mul(big.NewInt(50), pow18),
			price:      pow36,
			lltv:       big.NewInt(500000000000000000),
			want:       1.0,
		},
		{
			name:       "large_values",
			collateral: new(big.Int).Mul(pow18, big.NewInt(1_000_000)),
			debt:       big.NewInt(1),
			price:      pow36,
			lltv:       big.NewInt(1000000000000000000),
			want:       1e24,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, computeHealthFactor(tt.collateral, tt.debt, tt.price, tt.lltv), 1e-9)
		})
	}
}

func TestGetHealthFactor(t *testing.T) {
	pow18 := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	pow36 := new(big.Int).Exp(big.NewInt(10), big.NewInt(36), nil)
	lltv := big.NewInt(860000000000000000)

	r := NewHealthFactorReader()
	require.NotNil(t, r)

	tests := []struct {
		name       string
		collateral *big.Int
		debt       *big.Int
		price      *big.Int
		lltv       *big.Int
		want       float64
	}{
		{
			name:       "ok",
			collateral: pow18,
			debt:       pow18,
			price:      pow36,
			lltv:       lltv,
			want:       0.86,
		},
		{
			name:       "nil_debt",
			collateral: pow18,
			debt:       nil,
			price:      pow36,
			lltv:       lltv,
			want:       0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.GetHealthFactor(tt.collateral, tt.debt, tt.price, tt.lltv)
			assert.InDelta(t, tt.want, got, 1e-9)
		})
	}
}
