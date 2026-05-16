package morpho

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"test-task-edge/internal/types"
)

type stubScannerSuccess struct {
	ids []types.Bytes32
}

func (s *stubScannerSuccess) ScanMarkets(_ context.Context, _ uint64) ([]types.Bytes32, error) {
	return s.ids, nil
}

type stubScannerError struct{}

func (s *stubScannerError) ScanMarkets(_ context.Context, _ uint64) ([]types.Bytes32, error) {
	return nil, errors.New("scan err")
}

type stubScannerSequential struct {
	results [][]types.Bytes32
	call    int
}

func (s *stubScannerSequential) ScanMarkets(_ context.Context, _ uint64) ([]types.Bytes32, error) {
	if s.call >= len(s.results) {
		return nil, nil
	}
	r := s.results[s.call]
	s.call++
	return r, nil
}

func TestMarketProvider_cold_miss(t *testing.T) {
	ids := []types.Bytes32{{1}, {2}}
	p := NewMarketProvider(&stubScannerSuccess{ids: ids})

	got, err := p.DiscoverMarkets(context.Background(), 100)
	require.NoError(t, err)
	assert.ElementsMatch(t, ids, got)
}

func TestMarketProvider_cache_hit(t *testing.T) {
	p := NewMarketProvider(&stubScannerSequential{
		results: [][]types.Bytes32{{{1}}, {}},
	})

	got1, err := p.DiscoverMarkets(context.Background(), 100)
	require.NoError(t, err)
	assert.ElementsMatch(t, []types.Bytes32{{1}}, got1)

	got2, err := p.DiscoverMarkets(context.Background(), 200)
	require.NoError(t, err)
	assert.Equal(t, []types.Bytes32{{1}}, got2)
}

func TestMarketProvider_new_markets_accumulate(t *testing.T) {
	p := NewMarketProvider(&stubScannerSequential{
		results: [][]types.Bytes32{{{1}}, {{2}}},
	})

	got1, err := p.DiscoverMarkets(context.Background(), 100)
	require.NoError(t, err)
	assert.Equal(t, []types.Bytes32{{1}}, got1)

	got2, err := p.DiscoverMarkets(context.Background(), 200)
	require.NoError(t, err)
	assert.ElementsMatch(t, []types.Bytes32{{1}, {2}}, got2)
}

func TestMarketProvider_scan_error(t *testing.T) {
	p := NewMarketProvider(&stubScannerError{})
	_, err := p.DiscoverMarkets(context.Background(), 100)
	require.Error(t, err)
}
