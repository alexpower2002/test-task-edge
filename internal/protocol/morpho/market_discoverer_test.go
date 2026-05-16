package morpho

import (
	"context"
	"errors"
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

func (s *stubLogFiltererSuccess) FilterLogs(_ context.Context, _ ethereum.FilterQuery) ([]gethTypes.Log, error) {
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
			common.HexToHash(createMarketTopic),
			common.HexToHash("0x0100000000000000000000000000000000000000000000000000000000000000"),
		},
	}
	log2 := gethTypes.Log{
		Topics: []common.Hash{
			common.HexToHash(createMarketTopic),
			common.HexToHash("0x0200000000000000000000000000000000000000000000000000000000000000"),
		},
	}

	tests := []struct {
		name      string
		filterer  logFilterer
		fromBlock uint64
		latest    uint64
		want      []types.Bytes32
		wantErr   bool
	}{
		{
			name:      "single_market",
			filterer:  &stubLogFiltererSuccess{logs: []gethTypes.Log{log1}},
			fromBlock: 0,
			latest:    0,
			want:      []types.Bytes32{id1},
		},
		{
			name:      "multiple_markets",
			filterer:  &stubLogFiltererSuccess{logs: []gethTypes.Log{log1, log2}},
			fromBlock: 0,
			latest:    0,
			want:      []types.Bytes32{id1, id2},
		},
		{
			name:      "deduplication",
			filterer:  &stubLogFiltererSuccess{logs: []gethTypes.Log{log1, log1}},
			fromBlock: 0,
			latest:    0,
			want:      []types.Bytes32{id1},
		},
		{
			name:      "no_logs",
			filterer:  &stubLogFiltererSuccess{logs: nil},
			fromBlock: 0,
			latest:    0,
			want:      nil,
		},
		{
			name:      "filter_error",
			filterer:  &stubLogFiltererError{},
			fromBlock: 0,
			latest:    0,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewMarketDiscoverer(tt.filterer, contractAddr, tt.fromBlock)
			got, err := d.ScanMarkets(context.Background(), tt.latest)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
