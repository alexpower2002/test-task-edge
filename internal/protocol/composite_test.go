package protocol

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"test-task-edge/internal/types"
)

type mockParser struct {
	name      string
	positions []Position
	posErr    error
}

func (m *mockParser) ParsePositions(_ context.Context, _ []types.Address, _ types.BlockRef) ([]Position, error) {
	return m.positions, m.posErr
}

func TestComposite_ParsePositions(t *testing.T) {
	block := types.BlockRef{Number: 1, Timestamp: 100}
	wallets := []types.Address{{}}

	tests := []struct {
		name    string
		parsers []positionsParser
		wantPos int
	}{
		{
			name:    "empty_parsers",
			parsers: nil,
			wantPos: 0,
		},
		{
			name: "positions_collected",
			parsers: []positionsParser{
				&mockParser{
					name:      "p1",
					positions: []Position{{Protocol: "p1", BlockNumber: 1}},
				},
				&mockParser{
					name:      "p2",
					positions: []Position{{Protocol: "p2", BlockNumber: 1}},
				},
			},
			wantPos: 2,
		},
		{
			name: "error_skips_that_parser",
			parsers: []positionsParser{
				&mockParser{
					name:   "errParser",
					posErr: errors.New("fail"),
				},
				&mockParser{
					name:      "okParser",
					positions: []Position{{Protocol: "ok", BlockNumber: 1}},
				},
			},
			wantPos: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewComposite(tt.parsers...)
			positions, err := c.ParsePositions(context.Background(), wallets, block)
			require.NoError(t, err)
			require.Len(t, positions, tt.wantPos)
			for _, pos := range positions {
				assert.Equal(t, block.Number, pos.BlockNumber)
			}
		})
	}
}


