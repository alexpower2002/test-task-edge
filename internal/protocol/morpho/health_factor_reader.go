package morpho

import (
	"math/big"
)

const (
	morphoOraclePriceDecimals = 36
	morphoLLTVDecimals        = 18
)

type HealthFactorReader struct{}

func NewHealthFactorReader() *HealthFactorReader {
	return &HealthFactorReader{}
}

func (r *HealthFactorReader) GetHealthFactor(collateral, debt, price, lltv *big.Int) string {
	return computeHealthFactor(collateral, debt, price, lltv)
}

func computeHealthFactor(collateral, debt, price, lltv *big.Int) string {
	if debt == nil || debt.Sign() == 0 {
		return "0"
	}
	if collateral == nil || price == nil || lltv == nil {
		return "0"
	}
	num := new(big.Int).Mul(collateral, price)
	num.Mul(num, lltv)
	den := new(big.Int).Mul(debt, new(big.Int).Exp(big.NewInt(10), big.NewInt(morphoLLTVDecimals+morphoOraclePriceDecimals), nil))
	if den.Sign() == 0 {
		return "0"
	}
	return new(big.Rat).SetFrac(num, den).FloatString(6)
}
