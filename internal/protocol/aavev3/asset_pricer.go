package aavev3

import (
	"context"
	"fmt"
	"math/big"

	"test-task-edge/internal/types"
)

const selectorGetAssetPrice = "b3596f07"

type AssetPricer struct {
	client ethCaller
	oracle types.Address
}

func NewAssetPricer(client ethCaller, oracle types.Address) *AssetPricer {
	return &AssetPricer{client: client, oracle: oracle}
}

func (p *AssetPricer) GetAssetPrice(ctx context.Context, asset types.Address, block uint64) (*big.Int, error) {
	raw, err := p.client.EthCall(ctx, p.oracle, types.CallData(selectorGetAssetPrice, types.WordAddress(asset)), block)
	if err != nil {
		return nil, err
	}
	words, err := types.DecodeWords(raw)
	if err != nil {
		return nil, err
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("empty price response")
	}
	return types.WordBig(words[0])
}
