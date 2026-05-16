package aavev3

import (
	"context"
	"fmt"

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

func (r *HealthFactorReader) GetHealthFactor(ctx context.Context, user types.Address, block uint64) (float64, error) {
	raw, err := r.client.EthCall(ctx, r.pool, types.CallData(selectorGetUserAccountData, types.WordAddress(user)), block)
	if err != nil {
		return 0, err
	}
	words, err := types.DecodeWords(raw)
	if err != nil {
		return 0, err
	}
	return parseHealthFactor(words)
}

func parseHealthFactor(words []string) (float64, error) {
	if len(words) < 6 {
		return 0, fmt.Errorf("unexpected account data word count %d", len(words))
	}
	hf, err := types.WordBig(words[5])
	if err != nil {
		return 0, err
	}
	return utils.AmountToDecimal(hf, healthFactorDecimals), nil
}
