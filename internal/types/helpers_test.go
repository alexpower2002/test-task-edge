package types_test

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

func TestCallData(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		args     []string
		want     string
	}{
		{name: "selector_only", selector: "aabbccdd", want: "0xaabbccdd"},
		{name: "with_args", selector: "aabbccdd", args: []string{"eeff0011", "22334455"}, want: "0xaabbccddeeff001122334455"},
		{name: "no_args", selector: "aabbccdd", args: []string{}, want: "0xaabbccdd"},
		{name: "empty_selector", selector: "", want: "0x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.CallData(tt.selector, tt.args...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWordAddress(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		{
			name: "one",
			addr: "0x0000000000000000000000000000000000000001",
			want: "0000000000000000000000000000000000000000000000000000000000000001",
		},
		{
			name: "dead",
			addr: "0xdead000000000000000000000000000000000000",
			want: "000000000000000000000000dead000000000000000000000000000000000000",
		},
		{
			name: "zero",
			addr: "0x0000000000000000000000000000000000000000",
			want: "0000000000000000000000000000000000000000000000000000000000000000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := utils.ParseAddress(tt.addr)
			require.NoError(t, err)
			assert.Equal(t, tt.want, types.WordAddress(a))
		})
	}
}

func TestWordBytes32(t *testing.T) {
	makeB32 := func(s string) types.Bytes32 {
		raw, err := hex.DecodeString(s)
		require.NoError(t, err)
		var out types.Bytes32
		copy(out[:], raw)
		return out
	}
	tests := []struct {
		name string
		b    types.Bytes32
		want string
	}{
		{name: "zero", b: makeB32(strings.Repeat("00", 32)), want: strings.Repeat("00", 32)},
		{name: "ones", b: makeB32(strings.Repeat("11", 32)), want: strings.Repeat("11", 32)},
		{name: "max", b: makeB32(strings.Repeat("ff", 32)), want: strings.Repeat("ff", 32)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, types.WordBytes32(tt.b))
		})
	}
}

func TestDecodeWords(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{
			name: "single_word",
			raw:  "0x" + strings.Repeat("aa", 32),
			want: []string{strings.Repeat("aa", 32)},
		},
		{
			name: "two_words",
			raw:  "0x" + strings.Repeat("00", 32) + strings.Repeat("11", 32),
			want: []string{strings.Repeat("00", 32), strings.Repeat("11", 32)},
		},
		{
			name: "without_prefix",
			raw:  strings.Repeat("ff", 32),
			want: []string{strings.Repeat("ff", 32)},
		},
		{
			name:    "misaligned",
			raw:     "0xabc",
			wantErr: true,
		},
		{
			name:    "empty",
			raw:     "0x",
			want:    []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.DecodeWords(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWordBig(t *testing.T) {
	tests := []struct {
		name    string
		word    string
		want    string
		wantErr bool
	}{
		{name: "zero", word: "0000000000000000000000000000000000000000000000000000000000000000", want: "0"},
		{name: "one", word: "0000000000000000000000000000000000000000000000000000000000000001", want: "1"},
		{name: "max_uint256", word: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", want: "115792089237316195423570985008687907853269984665640564039457584007913129639935"},
		{name: "small", word: "000000000000000000000000000000000000000000000000000000000000000a", want: "10"},
		{name: "large", word: "0000000000000000000000000000000000000000000000000000000000aabbcc", want: "11189196"},
		{name: "mixed_case", word: "00000000000000000000000000000000000000000000000000000000000000Aa", want: "170"},
		{name: "invalid_hex", word: "zzz", wantErr: true},
		{name: "empty", word: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.WordBig(tt.word)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			want, ok := new(big.Int).SetString(tt.want, 10)
			require.True(t, ok)
			assert.Zero(t, got.Cmp(want))
		})
	}
}

func TestWordUint64(t *testing.T) {
	tests := []struct {
		name    string
		word    string
		want    uint64
		wantErr bool
	}{
		{name: "zero", word: "0000000000000000000000000000000000000000000000000000000000000000", want: 0},
		{name: "one", word: "0000000000000000000000000000000000000000000000000000000000000001", want: 1},
		{name: "max_uint64", word: "000000000000000000000000000000000000000000000000ffffffffffffffff", want: 18446744073709551615},
		{name: "overflow_truncated", word: "0000000000000000000000000000000000000000000000010000000000000000", want: 0},
		{name: "invalid_hex", word: "gg", wantErr: true},
		{name: "empty", word: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.WordUint64(tt.word)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWordToAddress(t *testing.T) {
	tests := []struct {
		name    string
		word    string
		want    string
		wantErr bool
	}{
		{
			name: "one",
			word: "0000000000000000000000000000000000000000000000000000000000000001",
			want: "0x0000000000000000000000000000000000000001",
		},
		{
			name: "dead",
			word: "000000000000000000000000dead000000000000000000000000000000000000",
			want: "0xdead000000000000000000000000000000000000",
		},
		{
			name: "zero",
			word: "0000000000000000000000000000000000000000000000000000000000000000",
			want: "0x0000000000000000000000000000000000000000",
		},
		{
			name:    "too_short",
			word:    "abcd",
			wantErr: true,
		},
		{
			name:    "too_long",
			word:    strings.Repeat("00", 33),
			wantErr: true,
		},
		{
			name:    "invalid_hex",
			word:    strings.Repeat("00", 12) + strings.Repeat("zz", 20),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.WordToAddress(tt.word)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestDecodeString(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{
			name: "utf8_eth",
			raw: "0x" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000003" +
				"4554480000000000000000000000000000000000000000000000000000000000",
			want: "ETH",
		},
		{
			name: "usdc",
			raw: "0x" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000004" +
				"5553444300000000000000000000000000000000000000000000000000000000",
			want: "USDC",
		},
		{
			name: "zero_length",
			raw: "0x" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000000",
			want: "",
		},
		{
			name:    "misaligned_hex",
			raw:     "0xabc",
			wantErr: true,
		},
		{
			name:    "empty_words",
			raw:     "0x",
			wantErr: true,
		},
		{
			name: "offset_beyond_words",
			raw: "0x" +
				"0000000000000000000000000000000000000000000000000000000000000060" +
				"0000000000000000000000000000000000000000000000000000000000000003" +
				"4554480000000000000000000000000000000000000000000000000000000000",
			wantErr: true,
		},
		{
			name: "size_exceeds_data",
			raw: "0x" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"0000000000000000000000000000000000000000000000000000000000000030" +
				"4554480000000000000000000000000000000000000000000000000000000000",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.DecodeString(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeStaticString(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{
			name: "usdc_padded",
			raw:  "0x5553444300000000000000000000000000000000000000000000000000000000",
			want: "USDC",
		},
		{
			name: "wbtc_padded",
			raw:  "0x5742544300000000000000000000000000000000000000000000000000000000",
			want: "WBTC",
		},
		{
			name: "no_padding",
			raw:  "0x55534443",
			want: "USDC",
		},
		{
			name: "empty_hex",
			raw:  "0x",
			want: "",
		},
		{
			name: "trailing_nonzero",
			raw:  "0x5553444301000000000000000000000000000000000000000000000000000000",
			want: "USDC\u0001",
		},
		{
			name:    "invalid_hex",
			raw:     "0xzz",
			wantErr: true,
		},
		{
			name: "without_prefix",
			raw:  "55534443",
			want: "USDC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.DecodeStaticString(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseAddress_roundTrip(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "one", raw: "0x0000000000000000000000000000000000000001"},
		{name: "dead", raw: "0xdead000000000000000000000000000000000000"},
		{name: "zero", raw: "0x0000000000000000000000000000000000000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := utils.ParseAddress(tt.raw)
			require.NoError(t, err)
			assert.Equal(t, tt.raw, addr.String())
			got, err := types.WordToAddress(types.WordAddress(addr))
			require.NoError(t, err)
			assert.Equal(t, tt.raw, got.String())
		})
	}
}
