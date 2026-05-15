package aavev3

import (
	"context"
	"fmt"
	"math/big"

	"test-task-edge/internal/types"
)

const selectorGetUserReserveData = "28dd2d01"

type UserDataReader struct {
	client       ethCaller
	dataProvider types.Address
}

func NewUserDataReader(client ethCaller, dataProvider types.Address) *UserDataReader {
	return &UserDataReader{client: client, dataProvider: dataProvider}
}

func (r *UserDataReader) GetUserState(ctx context.Context, asset, user types.Address, block uint64) (deposit *big.Int, borrow *big.Int, err error) {
	raw, err := r.client.EthCall(ctx, r.dataProvider, types.CallData(selectorGetUserReserveData, types.WordAddress(asset), types.WordAddress(user)), block)
	if err != nil {
		return nil, nil, err
	}
	words, err := types.DecodeWords(raw)
	if err != nil {
		return nil, nil, err
	}
	return parseUserState(words)
}

func parseUserState(words []string) (deposit *big.Int, borrow *big.Int, err error) {
	if len(words) < 3 {
		return nil, nil, fmt.Errorf("unexpected getUserReserveData word count %d", len(words))
	}
	aToken, err := types.WordBig(words[0])
	if err != nil {
		return nil, nil, err
	}
	variableDebt, err := types.WordBig(words[2])
	if err != nil {
		return nil, nil, err
	}
	return aToken, variableDebt, nil
}
