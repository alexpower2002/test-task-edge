package morpho

import (
	"context"
	"fmt"
	"math/big"

	"test-task-edge/internal/types"
)

const selectorMorphoMarketParams = "2c3c9157" // marketParams(bytes32)

type MarketParamsReader struct {
	client  ethCaller
	address types.Address
}

func NewMarketParamsReader(client ethCaller, address types.Address) *MarketParamsReader {
	return &MarketParamsReader{client: client, address: address}
}

func (r *MarketParamsReader) GetMarketConfig(ctx context.Context, id types.Bytes32, block uint64) (loanToken types.Address, collateralToken types.Address, oracle types.Address, lltv *big.Int, err error) {
	raw, err := r.client.EthCall(ctx, r.address, types.CallData(selectorMorphoMarketParams, types.WordBytes32(id)), block)
	if err != nil {
		return types.Address{}, types.Address{}, types.Address{}, nil, err
	}
	words, err := types.DecodeWords(raw)
	if err != nil {
		return types.Address{}, types.Address{}, types.Address{}, nil, err
	}
	return parseMarketConfig(words)
}

func parseMarketConfig(words []string) (loanToken types.Address, collateralToken types.Address, oracle types.Address, lltv *big.Int, err error) {
	if len(words) < 5 {
		return types.Address{}, types.Address{}, types.Address{}, nil, fmt.Errorf("unexpected market params word count %d", len(words))
	}
	loan, err := types.WordToAddress(words[0])
	if err != nil {
		return types.Address{}, types.Address{}, types.Address{}, nil, err
	}
	collateral, err := types.WordToAddress(words[1])
	if err != nil {
		return types.Address{}, types.Address{}, types.Address{}, nil, err
	}
	oracle, err = types.WordToAddress(words[2])
	if err != nil {
		return types.Address{}, types.Address{}, types.Address{}, nil, err
	}
	lltv, err = types.WordBig(words[4])
	if err != nil {
		return types.Address{}, types.Address{}, types.Address{}, nil, err
	}
	return loan, collateral, oracle, lltv, nil
}
