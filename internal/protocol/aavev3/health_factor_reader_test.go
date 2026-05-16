package aavev3

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"test-task-edge/internal/types"
	"test-task-edge/internal/utils"
)

type stubHealthSuccess struct {
	result string
}

func (s *stubHealthSuccess) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return s.result, nil
}

type stubHealthError struct{}

func (s *stubHealthError) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return "", errors.New("rpc err")
}

func word(h string) string {
	if len(h) > 64 {
		return h[:64]
	}
	return strings.Repeat("0", 64-len(h)) + h
}

func TestParseHealthFactor(t *testing.T) {
	hfValue := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	hfWord := word(hfValue.Text(16))

	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	maxWord := word(max.Text(16))

	makeWords := func(h string, n int) []string {
		w := make([]string, n)
		for i := range w {
			w[i] = strings.Repeat("0", 64)
		}
		if n > 5 {
			w[5] = h
		}
		return w
	}

	tests := []struct {
		name    string
		words   []string
		want    float64
		wantErr bool
	}{
		{
			name:  "ok",
			words: makeWords(hfWord, 6),
			want:  1.0,
		},
		{
			name:  "small_value",
			words: makeWords(word(new(big.Int).SetInt64(1).Text(16)), 6),
			want:  1e-18,
		},
		{
			name:  "max_uint256",
			words: makeWords(maxWord, 6),
			want:  0,
		},
		{
			name:    "too_few_words",
			words:   makeWords(hfWord, 3),
			wantErr: true,
		},
		{
			name:  "more_than_6_words",
			words: makeWords(hfWord, 10),
			want:  1.0,
		},
		{
			name:    "invalid_word",
			words:   []string{strings.Repeat("0", 64), strings.Repeat("0", 64), strings.Repeat("0", 64), strings.Repeat("0", 64), strings.Repeat("0", 64), "zz"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHealthFactor(tt.words)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.InDelta(t, tt.want, got, 1e-12)
		})
	}
}

func TestGetHealthFactor(t *testing.T) {
	pool, err := utils.ParseAddress("0x00000000000000000000000000000000000000aa")
	require.NoError(t, err)
	user, err := utils.ParseAddress("0x0000000000000000000000000000000000000001")
	require.NoError(t, err)

	hfValue := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	hfWord := word(hfValue.Text(16))

	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	maxWord := word(max.Text(16))

	makeResponse := func(h string) string {
		return "0x" + strings.Repeat("00", 32*5) + h
	}

	tests := []struct {
		name    string
		caller  ethCaller
		want    float64
		wantErr bool
	}{
		{
			name:   "ok",
			caller: &stubHealthSuccess{result: makeResponse(hfWord)},
			want:   1.0,
		},
		{
			name: "small_value",
			caller: &stubHealthSuccess{result: makeResponse(
				word(new(big.Int).SetInt64(1).Text(16)),
			)},
			want: 1e-18,
		},
		{
			name:    "eth_call_error",
			caller:  &stubHealthError{},
			wantErr: true,
		},
		{
			name:    "decode_words_error",
			caller:  &stubHealthSuccess{result: "0xabc"},
			wantErr: true,
		},
		{
			name:    "too_few_words",
			caller:  &stubHealthSuccess{result: "0x" + strings.Repeat("00", 32*3)},
			wantErr: true,
		},
		{
			name:   "max_uint256",
			caller: &stubHealthSuccess{result: makeResponse(maxWord)},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewHealthFactorReader(tt.caller, pool)
			got, err := r.GetHealthFactor(context.Background(), user, 100)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.InDelta(t, tt.want, got, 1e-12)
		})
	}
}
