package ethereum

import (
	"context"

	"test-task-edge/internal/types"
)

const (
	selectorSymbol   = "95d89b41"
	selectorDecimals = "313ce567"
)

type ethCaller interface {
	EthCall(ctx context.Context, to types.Address, data string, block uint64) (string, error)
}

type TokenReader struct {
	client ethCaller
}

func NewTokenReader(client ethCaller) *TokenReader {
	return &TokenReader{client: client}
}

func (r *TokenReader) ReadToken(ctx context.Context, token types.Address, block uint64) types.Token {
	item := types.Token{Address: token, Symbol: token.String(), Decimals: 18}
	if raw, err := r.client.EthCall(ctx, token, types.CallData(selectorDecimals), block); err == nil {
		if words, err := types.DecodeWords(raw); err == nil && len(words) > 0 {
			if n, err := types.WordUint64(words[0]); err == nil {
				item.Decimals = int(n)
			}
		}
	}
	if raw, err := r.client.EthCall(ctx, token, types.CallData(selectorSymbol), block); err == nil {
		if symbol, err := types.DecodeString(raw); err == nil {
			item.Symbol = symbol
		} else if symbol, err := types.DecodeStaticString(raw); err == nil && symbol != "" {
			item.Symbol = symbol
		}
	}
	return item
}
