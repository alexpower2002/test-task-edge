package morpho

import (
	"context"
	"fmt"
	"math/big"

	"test-task-edge/internal/types"
)

const selectorMorphoMarket = "5c60e39a" // market(bytes32)

type MarketDataReader struct {
	client  ethCaller
	address types.Address
}

func NewMarketDataReader(client ethCaller, address types.Address) *MarketDataReader {
	return &MarketDataReader{client: client, address: address}
}

func (r *MarketDataReader) GetMarketState(ctx context.Context, id types.Bytes32, block uint64) (totalBorrowAssets *big.Int, totalBorrowShares *big.Int, err error) {
	raw, err := r.client.EthCall(ctx, r.address, types.CallData(selectorMorphoMarket, types.WordBytes32(id)), block)
	if err != nil {
		return nil, nil, err
	}
	words, err := types.DecodeWords(raw)
	if err != nil {
		return nil, nil, err
	}
	return parseMarketState(words)
}

func parseMarketState(words []string) (totalBorrowAssets *big.Int, totalBorrowShares *big.Int, err error) {
	if len(words) < 4 {
		return nil, nil, fmt.Errorf("unexpected market word count %d", len(words))
	}
	totalBorrowAssets, err = types.WordBig(words[2])
	if err != nil {
		return nil, nil, err
	}
	totalBorrowShares, err = types.WordBig(words[3])
	if err != nil {
		return nil, nil, err
	}
	return totalBorrowAssets, totalBorrowShares, nil
}
