package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"test-task-edge/internal/types"
)

func ParseAddress(raw string) (types.Address, error) {
	var out types.Address
	raw = strings.TrimPrefix(strings.TrimSpace(raw), "0x")
	if len(raw) != 40 {
		return out, fmt.Errorf("invalid address length: %q", raw)
	}
	data, err := hex.DecodeString(raw)
	if err != nil {
		return out, err
	}
	copy(out[:], data)
	return out, nil
}

func SharesToAssets(shares, totalAssets, totalShares *big.Int) *big.Int {
	if shares == nil || totalAssets == nil || totalShares == nil || shares.Sign() == 0 || totalShares.Sign() == 0 {
		return big.NewInt(0)
	}
	return new(big.Int).Div(new(big.Int).Mul(shares, totalAssets), totalShares)
}

func DebtTokenName(symbol string, debt *big.Int) string {
	if debt == nil || debt.Sign() == 0 {
		return ""
	}
	return symbol
}

func AmountToDecimal(n *big.Int, decimals int) float64 {
	if n == nil {
		return 0
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	f, _ := new(big.Rat).SetFrac(n, scale).Float64()
	return f
}

func ParseAddressList(raw []string) ([]types.Address, error) {
	out := make([]types.Address, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		addr, err := ParseAddress(s)
		if err != nil {
			return nil, err
		}
		out = append(out, addr)
	}
	return out, nil
}
