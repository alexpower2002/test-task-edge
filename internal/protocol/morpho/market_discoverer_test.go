package morpho

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

type stubLogFiltererSuccess struct {
	logs []gethTypes.Log
}

func (s *stubLogFiltererSuccess) FilterLogs(_ context.Context, q ethereum.FilterQuery) ([]gethTypes.Log, error) {
	return s.logs, nil
}

type stubLogFiltererAssert struct {
	wantFrom *big.Int
	wantTo   *big.Int
	logs     []gethTypes.Log
}

func (s *stubLogFiltererAssert) FilterLogs(_ context.Context, q ethereum.FilterQuery) ([]gethTypes.Log, error) {
	if q.FromBlock.Cmp(s.wantFrom) != 0 {
		return nil, errors.New("unexpected from block")
	}
	if q.ToBlock.Cmp(s.wantTo) != 0 {
		return nil, errors.New("unexpected to block")
	}
	return s.logs, nil
}

type stubLogFiltererError struct{}

func (s *stubLogFiltererError) FilterLogs(_ context.Context, _ ethereum.FilterQuery) ([]gethTypes.Log, error) {
	return nil, errors.New("filter err")
}

func bytes32FromHash(t *testing.T, hex string) types.Bytes32 {
	t.Helper()
	h := common.HexToHash(hex)
	var b types.Bytes32
	copy(b[:], h.Bytes())
	return b
}

func TestScanMarkets(t *testing.T) {
	contractAddr, err := utils.ParseAddress("0x00000000000000000000000000000000000000aa")
	require.NoError(t, err)

	id1 := bytes32FromHash(t, "0x0100000000000000000000000000000000000000000000000000000000000000")
	id2 := bytes32FromHash(t, "0x0200000000000000000000000000000000000000000000000000000000000000")

	log1 := gethTypes.Log{
		Topics: []common.Hash{
			common.HexToHash("0xac4b2400f169220b0c0afdde7a0b32e775ba727ea1cb30b35f935cdaab8683ac"),
			common.HexToHash("0x0100000000000000000000000000000000000000000000000000000000000000"),
		},
	}
	log2 := gethTypes.Log{
		Topics: []common.Hash{
			common.HexToHash("0xac4b2400f169220b0c0afdde7a0b32e775ba727ea1cb30b35f935cdaab8683ac"),
			common.HexToHash("0x0200000000000000000000000000000000000000000000000000000000000000"),
		},
	}

	tests := []struct {
		name        string
		filterer    logFilterer
		deployBlock uint64
		blockNumber uint64
		want        []types.Bytes32
		wantErr     bool
	}{
		{
			name:        "single_market",
			filterer:    &stubLogFiltererAssert{wantFrom: big.NewInt(0), wantTo: big.NewInt(42), logs: []gethTypes.Log{log1}},
			deployBlock: 0,
			blockNumber: 42,
			want:        []types.Bytes32{id1},
		},
		{
			name:        "multiple_markets",
			filterer:    &stubLogFiltererAssert{wantFrom: big.NewInt(0), wantTo: big.NewInt(99), logs: []gethTypes.Log{log1, log2}},
			deployBlock: 0,
			blockNumber: 99,
			want:        []types.Bytes32{id1, id2},
		},
		{
			name:        "deduplication",
			filterer:    &stubLogFiltererSuccess{logs: []gethTypes.Log{log1, log1}},
			deployBlock: 0,
			blockNumber: 100,
			want:        []types.Bytes32{id1},
		},
		{
			name:        "no_logs",
			filterer:    &stubLogFiltererSuccess{logs: nil},
			deployBlock: 0,
			blockNumber: 100,
			want:        nil,
		},
		{
			name:        "filter_error",
			filterer:    &stubLogFiltererError{},
			deployBlock: 0,
			blockNumber: 100,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewMarketDiscoverer(tt.filterer, contractAddr, tt.deployBlock, 10000)
			got, err := d.ScanMarkets(context.Background(), tt.blockNumber)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
