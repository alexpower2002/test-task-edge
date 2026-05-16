package morpho

import (
	"context"
	"fmt"
	"math/big"

	"test-task-edge/internal/types"
)

const selectorMorphoOraclePrice = "a035b1fe" // price()

type OraclePricer struct {
	client ethCaller
}

func NewOraclePricer(client ethCaller) *OraclePricer {
	return &OraclePricer{client: client}
}

func (p *OraclePricer) GetOraclePrice(ctx context.Context, oracle types.Address, block uint64) (*big.Int, error) {
	if oracle.IsZero() {
		return big.NewInt(0), nil
	}
	raw, err := p.client.EthCall(ctx, oracle, types.CallData(selectorMorphoOraclePrice), block)
	if err != nil {
		return nil, err
	}
	words, err := types.DecodeWords(raw)
	if err != nil {
		return nil, err
	}
	return parseOraclePrice(words)
}

func parseOraclePrice(words []string) (*big.Int, error) {
	if len(words) == 0 {
		return nil, fmt.Errorf("empty oracle price response")
	}
	return types.WordBig(words[0])
}
