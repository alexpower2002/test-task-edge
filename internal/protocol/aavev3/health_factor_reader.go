package aavev3

import (
	"context"
	"fmt"
	"math/big"

	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

const (
	selectorGetUserAccountData = "bf92857c"
	healthFactorDecimals       = 18
)

type HealthFactorReader struct {
	client ethCaller
	pool   types.Address
}

func NewHealthFactorReader(client ethCaller, pool types.Address) *HealthFactorReader {
	return &HealthFactorReader{client: client, pool: pool}
}

func (r *HealthFactorReader) GetHealthFactor(ctx context.Context, user types.Address, block uint64) (string, error) {
	raw, err := r.client.EthCall(ctx, r.pool, types.CallData(selectorGetUserAccountData, types.WordAddress(user)), block)
	if err != nil {
		return "", err
	}
	words, err := types.DecodeWords(raw)
	if err != nil {
		return "", err
	}
	if len(words) < 6 {
		return "", fmt.Errorf("unexpected account data word count %d", len(words))
	}
	hf, err := types.WordBig(words[5])
	if err != nil {
		return "", err
	}
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	if hf.Cmp(max) == 0 {
		return "0", nil
	}
	return utils.AmountToDecimal(hf, healthFactorDecimals), nil
}
