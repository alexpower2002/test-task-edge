package morpho

import (
	"context"
	"fmt"
	"math/big"

	"test-task-edge/internal/types"
)

const selectorMorphoPosition = "93c52062" // position(bytes32,address)

type ethCaller interface {
	EthCall(ctx context.Context, to types.Address, data string, block uint64) (string, error)
}

type PositionReader struct {
	client  ethCaller
	address types.Address
}

func NewPositionReader(client ethCaller, address types.Address) *PositionReader {
	return &PositionReader{client: client, address: address}
}

func (r *PositionReader) GetUserPosition(ctx context.Context, id types.Bytes32, user types.Address, block uint64) (supplyShares *big.Int, borrowShares *big.Int, collateral *big.Int, err error) {
	raw, err := r.client.EthCall(ctx, r.address, types.CallData(selectorMorphoPosition, types.WordBytes32(id), types.WordAddress(user)), block)
	if err != nil {
		return nil, nil, nil, err
	}
	words, err := types.DecodeWords(raw)
	if err != nil {
		return nil, nil, nil, err
	}
	return parsePositionData(words)
}

func parsePositionData(words []string) (supplyShares *big.Int, borrowShares *big.Int, collateral *big.Int, err error) {
	if len(words) < 3 {
		return nil, nil, nil, fmt.Errorf("unexpected position word count %d", len(words))
	}
	supplyShares, err = types.WordBig(words[0])
	if err != nil {
		return nil, nil, nil, err
	}
	borrowShares, err = types.WordBig(words[1])
	if err != nil {
		return nil, nil, nil, err
	}
	collateral, err = types.WordBig(words[2])
	if err != nil {
		return nil, nil, nil, err
	}
	return supplyShares, borrowShares, collateral, nil
}
