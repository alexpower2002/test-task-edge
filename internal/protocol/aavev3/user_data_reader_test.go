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

type stubUserDataSuccess struct {
	result string
}

func (s *stubUserDataSuccess) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return s.result, nil
}

type stubUserDataError struct{}

func (s *stubUserDataError) EthCall(_ context.Context, _ types.Address, _ string, _ uint64) (string, error) {
	return "", errors.New("rpc err")
}

func TestParseUserState(t *testing.T) {
	one := big.NewInt(1)

	bigVal := new(big.Int).Lsh(big.NewInt(1), 200)
	bigWord := word(bigVal.Text(16))

	makeWords := func(words ...string) []string {
		return words
	}

	tests := []struct {
		name         string
		words        []string
		wantDeposit  *big.Int
		wantBorrow   *big.Int
		wantErr      bool
	}{
		{
			name:        "ok",
			words:       makeWords(word("1"), strings.Repeat("0", 64), word("2")),
			wantDeposit: one,
			wantBorrow:  big.NewInt(2),
		},
		{
			name:    "too_few_words",
			words:   makeWords(word("1"), word("2")),
			wantErr: true,
		},
		{
			name:    "empty",
			words:   nil,
			wantErr: true,
		},
		{
			name:    "invalid_deposit",
			words:   makeWords("zz", strings.Repeat("0", 64), word("2")),
			wantErr: true,
		},
		{
			name:    "invalid_borrow",
			words:   makeWords(word("1"), strings.Repeat("0", 64), "zz"),
			wantErr: true,
		},
		{
			name:        "zero_values",
			words:       makeWords(strings.Repeat("0", 64), strings.Repeat("0", 64), strings.Repeat("0", 64)),
			wantDeposit: new(big.Int),
			wantBorrow:  new(big.Int),
		},
		{
			name:        "large_values",
			words:       makeWords(bigWord, strings.Repeat("0", 64), bigWord),
			wantDeposit: bigVal,
			wantBorrow:  bigVal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deposit, borrow, err := parseUserState(tt.words)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Zero(t, tt.wantDeposit.Cmp(deposit))
			assert.Zero(t, tt.wantBorrow.Cmp(borrow))
		})
	}
}

func TestGetUserState(t *testing.T) {
	provider, err := utils.ParseAddress("0x00000000000000000000000000000000000000dd")
	require.NoError(t, err)
	asset, err := utils.ParseAddress("0x00000000000000000000000000000000000000aa")
	require.NoError(t, err)
	user, err := utils.ParseAddress("0x0000000000000000000000000000000000000001")
	require.NoError(t, err)

	makeResponse := func(words ...string) string {
		return "0x" + strings.Join(words, "")
	}

	tests := []struct {
		name        string
		caller      ethCaller
		wantDeposit *big.Int
		wantBorrow  *big.Int
		wantErr     bool
	}{
		{
			name:        "ok",
			caller:      &stubUserDataSuccess{result: makeResponse(word("5"), strings.Repeat("0", 64), word("3"))},
			wantDeposit: big.NewInt(5),
			wantBorrow:  big.NewInt(3),
		},
		{
			name:    "eth_call_error",
			caller:  &stubUserDataError{},
			wantErr: true,
		},
		{
			name:    "decode_words_error",
			caller:  &stubUserDataSuccess{result: "0xabc"},
			wantErr: true,
		},
		{
			name:    "too_few_words",
			caller:  &stubUserDataSuccess{result: "0x" + strings.Repeat("00", 32*2)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewUserDataReader(tt.caller, provider)
			deposit, borrow, err := r.GetUserState(context.Background(), asset, user, 100)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Zero(t, tt.wantDeposit.Cmp(deposit))
			assert.Zero(t, tt.wantBorrow.Cmp(borrow))
		})
	}
}
