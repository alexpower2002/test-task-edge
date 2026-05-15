package aavev3

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

type stubDiscoverSuccess struct {
	result string
}

func (s *stubDiscoverSuccess) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return s.result, nil
}

type stubDiscoverError struct{}

func (s *stubDiscoverError) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return "", errors.New("rpc err")
}

func makeReservesResponse(addrHexes ...string) string {
	count := len(addrHexes)
	offset := uint64(32)
	base := int(offset / 32)
	tupleStart := base + 1 + count

	totalWords := tupleStart + count*2 + 1
	words := make([]string, totalWords)
	words[0] = fmt.Sprintf("%064x", offset)
	words[base] = fmt.Sprintf("%064x", count)
	for i, h := range addrHexes {
		idx := tupleStart + i*2 + 1
		words[idx] = h
	}
	for i := range words {
		if words[i] == "" {
			words[i] = strings.Repeat("0", 64)
		}
	}
	return "0x" + strings.Join(words, "")
}

func TestParseReserveMetadata(t *testing.T) {
	tests := []struct {
		name         string
		words        []string
		wantCount    int
		wantStart    int
		wantErr      bool
	}{
		{
			name:      "two_reserves",
			words:     []string{"0000000000000000000000000000000000000000000000000000000000000020", "0000000000000000000000000000000000000000000000000000000000000002", strings.Repeat("0", 64)},
			wantCount: 2,
			wantStart: 4,
		},
		{
			name:      "single_reserve",
			words:     []string{"0000000000000000000000000000000000000000000000000000000000000020", "0000000000000000000000000000000000000000000000000000000000000001", strings.Repeat("0", 64)},
			wantCount: 1,
			wantStart: 3,
		},
		{
			name:    "too_few_words",
			words:   []string{"0000000000000000000000000000000000000000000000000000000000000020"},
			wantErr: true,
		},
		{
			name:    "empty",
			words:   nil,
			wantErr: true,
		},
		{
			name:    "invalid_offset",
			words:   []string{"zz", "0000000000000000000000000000000000000000000000000000000000000001"},
			wantErr: true,
		},
		{
			name:    "offset_beyond_words",
			words:   []string{"0000000000000000000000000000000000000000000000000000000000000060", "0000000000000000000000000000000000000000000000000000000000000001"},
			wantErr: true,
		},
		{
			name:    "invalid_count",
			words:   []string{"0000000000000000000000000000000000000000000000000000000000000020", "zz"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, start, err := parseReserveMetadata(tt.words)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
			assert.Equal(t, tt.wantStart, start)
		})
	}
}

func TestCollectReserveAddresses(t *testing.T) {
	addr1 := address(t, "0x00000000000000000000000000000000000000aa")
	addr2 := address(t, "0x00000000000000000000000000000000000000bb")

	tests := []struct {
		name       string
		words      []string
		count      int
		tupleStart int
		want       []types.Address
	}{
		{
			name:       "two_addresses",
			words:      []string{"", "", "", "", types.WordAddress(addr1), "", types.WordAddress(addr2)},
			count:      2,
			tupleStart: 3,
			want:       []types.Address{addr1, addr2},
		},
		{
			name:       "single_address",
			words:      []string{"", "", "", "", types.WordAddress(addr1)},
			count:      1,
			tupleStart: 3,
			want:       []types.Address{addr1},
		},
		{
			name:       "zero_address",
			words:      []string{"", "", "", "", types.WordAddress(types.Address{})},
			count:      1,
			tupleStart: 3,
			want:       []types.Address{},
		},
		{
			name:       "idx_out_of_bounds",
			words:      []string{"", ""},
			count:      1,
			tupleStart: 3,
			want:       []types.Address{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectReserveAddresses(tt.words, tt.count, tt.tupleStart)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiscoverReserves(t *testing.T) {
	provider, err := utils.ParseAddress("0x00000000000000000000000000000000000000dd")
	require.NoError(t, err)

	reserve1 := address(t, "0x00000000000000000000000000000000000000aa")
	reserve2 := address(t, "0x00000000000000000000000000000000000000bb")

	tests := []struct {
		name    string
		caller  ethCaller
		want    []types.Address
		wantErr bool
	}{
		{
			name:   "two_reserves",
			caller: &stubDiscoverSuccess{result: makeReservesResponse(types.WordAddress(reserve1), types.WordAddress(reserve2))},
			want:   []types.Address{reserve1, reserve2},
		},
		{
			name:   "single_reserve",
			caller: &stubDiscoverSuccess{result: makeReservesResponse(types.WordAddress(reserve1))},
			want:   []types.Address{reserve1},
		},
		{
			name:    "eth_call_error",
			caller:  &stubDiscoverError{},
			wantErr: true,
		},
		{
			name:    "decode_words_error",
			caller:  &stubDiscoverSuccess{result: "0xabc"},
			wantErr: true,
		},
		{
			name:    "too_few_words",
			caller:  &stubDiscoverSuccess{result: "0x" + strings.Repeat("00", 32)},
			wantErr: true,
		},
		{
			name:    "all_zero_addresses",
			caller:  &stubDiscoverSuccess{result: makeReservesResponse(types.WordAddress(types.Address{}))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewReserveDiscoverer(tt.caller, provider)
			got, err := d.DiscoverReserves(context.Background(), 100)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
