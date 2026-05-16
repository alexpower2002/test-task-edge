package ethereum

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"test-task-edge/internal/types"
)

type stubTokenReader struct {
	token types.Token
}

func (s *stubTokenReader) ReadToken(_ context.Context, _ types.Address, _ uint64) types.Token {
	return s.token
}

type trackingTokenReader struct {
	token     types.Token
	callCount int
}

func (s *trackingTokenReader) ReadToken(_ context.Context, _ types.Address, _ uint64) types.Token {
	s.callCount++
	return s.token
}

func TestProvider_GetToken(t *testing.T) {
	addr := address(t, "0x0000000000000000000000000000000000000001")
	token := types.Token{Address: addr, Symbol: "USDC", Decimals: 6}

	p := NewTokenProvider(&stubTokenReader{token: token})
	got := p.GetToken(context.Background(), addr, 100)
	assert.Equal(t, token, got)
}

func TestProvider_caching(t *testing.T) {
	addr := address(t, "0x0000000000000000000000000000000000000001")
	token := types.Token{Address: addr, Symbol: "USDC", Decimals: 6}

	reader := &trackingTokenReader{token: token}
	p := NewTokenProvider(reader)

	got1 := p.GetToken(context.Background(), addr, 100)
	assert.Equal(t, "USDC", got1.Symbol)

	got2 := p.GetToken(context.Background(), addr, 200)
	assert.Equal(t, "USDC", got2.Symbol)
	assert.Equal(t, got1, got2)

	assert.Equal(t, 1, reader.callCount)
}
