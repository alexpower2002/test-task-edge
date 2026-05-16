package ethereum

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

type mockEthClient struct {
	mock.Mock
}

func (m *mockEthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*gethTypes.Header, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gethTypes.Header), args.Error(1)
}

func (m *mockEthClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, msg, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockEthClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]gethTypes.Log, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]gethTypes.Log), args.Error(1)
}

func (m *mockEthClient) Close() {
	_ = m.Called()
}

func TestLatestBlock(t *testing.T) {
	tests := []struct {
		name    string
		header  *gethTypes.Header
		err     error
		want    types.BlockRef
		wantErr bool
	}{
		{
			name:   "ok",
			header: &gethTypes.Header{Number: big.NewInt(42), Time: 1000},
			want:   types.BlockRef{Number: 42, Timestamp: 1000},
		},
		{
			name:    "error",
			header:  nil,
			err:     assert.AnError,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockEthClient)
			m.On("HeaderByNumber", mock.Anything, mock.Anything).Return(tt.header, tt.err)

			client := NewRPCClientWithClient(m, time.Second)
			got, err := client.LatestBlock(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBlockByNumber(t *testing.T) {
	tests := []struct {
		name    string
		header  *gethTypes.Header
		err     error
		want    types.BlockRef
		wantErr bool
	}{
		{
			name:   "ok",
			header: &gethTypes.Header{Number: big.NewInt(99), Time: 2000},
			want:   types.BlockRef{Number: 99, Timestamp: 2000},
		},
		{
			name:    "error",
			header:  nil,
			err:     assert.AnError,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockEthClient)
			m.On("HeaderByNumber", mock.Anything, big.NewInt(99)).Return(tt.header, tt.err)

			client := NewRPCClientWithClient(m, time.Second)
			got, err := client.BlockByNumber(context.Background(), 99)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEthCall(t *testing.T) {
	to, err := utils.ParseAddress("0x0000000000000000000000000000000000000001")
	require.NoError(t, err)

	tests := []struct {
		name    string
		data    string
		block   uint64
		result  []byte
		err     error
		want    string
		wantErr bool
	}{
		{
			name:   "ok",
			data:   "0x01",
			block:  100,
			result: []byte{0xde, 0xad},
			want:   "0xdead",
		},
		{
			name:    "error",
			data:    "0x01",
			block:   0,
			err:     assert.AnError,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockEthClient)

			var blockNum *big.Int
			if tt.block > 0 {
				blockNum = new(big.Int).SetUint64(tt.block)
			}

			expectedMsg := ethereum.CallMsg{
				To:   (*common.Address)(&to),
				Data: common.FromHex(tt.data),
			}
			m.On("CallContract", mock.Anything, expectedMsg, blockNum).Return(tt.result, tt.err)

			client := NewRPCClientWithClient(m, time.Second)
			got, err := client.EthCall(context.Background(), to, tt.data, tt.block)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLogs(t *testing.T) {
	addr, err := utils.ParseAddress("0x0000000000000000000000000000000000000001")
	require.NoError(t, err)

	tests := []struct {
		name    string
		from    uint64
		to      uint64
		topics  []string
		logs    []gethTypes.Log
		err     error
		wantLen int
		wantErr bool
	}{
		{
			name:   "ok_no_topics",
			from:   1,
			to:     2,
			logs: []gethTypes.Log{
				{Address: common.Address(addr), BlockNumber: 1, Index: 0, TxHash: common.HexToHash("0xaa")},
			},
			wantLen: 1,
		},
		{
			name:   "ok_with_topics",
			from:   5,
			to:     5,
			topics: []string{"0x0000000000000000000000000000000000000000000000000000000000000001"},
			logs: []gethTypes.Log{
				{
					Address:     common.Address(addr),
					BlockNumber: 5,
					Index:       1,
					TxHash:      common.HexToHash("0xbb"),
					Topics:      []common.Hash{common.HexToHash("0x01")},
					Data:        []byte{0xff},
				},
			},
			wantLen: 1,
		},
		{
			name:    "error",
			from:    1,
			to:      1,
			err:     assert.AnError,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockEthClient)

			expectedQuery := ethereum.FilterQuery{
				FromBlock: new(big.Int).SetUint64(tt.from),
				ToBlock:   new(big.Int).SetUint64(tt.to),
				Addresses: []common.Address{common.Address(addr)},
			}
			if len(tt.topics) > 0 {
				hashes := make([]common.Hash, len(tt.topics))
				for i, tp := range tt.topics {
					hashes[i] = common.HexToHash(tp)
				}
				expectedQuery.Topics = [][]common.Hash{hashes}
			}

			m.On("FilterLogs", mock.Anything, expectedQuery).Return(tt.logs, tt.err)

			client := NewRPCClientWithClient(m, time.Second)
			got, err := client.Logs(context.Background(), addr, tt.from, tt.to, tt.topics)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
		})
	}
}
