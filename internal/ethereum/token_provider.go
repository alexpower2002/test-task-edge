package ethereum

import (
	"context"
	"sync"

	"test-task-edge/internal/types"
)

type tokenReader interface {
	ReadToken(ctx context.Context, token types.Address, block uint64) types.Token
}

type TokenProvider struct {
	reader tokenReader
	cache  sync.Map
}

func NewTokenProvider(reader tokenReader) *TokenProvider {
	return &TokenProvider{reader: reader}
}

func (p *TokenProvider) GetToken(ctx context.Context, token types.Address, block uint64) types.Token {
	if item, ok := p.cache.Load(token); ok {
		return item.(types.Token)
	}

	item := p.reader.ReadToken(ctx, token, block)
	p.cache.Store(token, item)
	return item
}
