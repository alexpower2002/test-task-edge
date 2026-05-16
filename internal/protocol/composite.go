package protocol

import (
	"context"

	"test-task-edge/internal/types"
)

type positionsParser interface {
	ParsePositions(ctx context.Context, wallets []types.Address, block types.BlockRef) ([]Position, error)
}

type Composite struct {
	positionsParsers []positionsParser
}

func NewComposite(parsers ...positionsParser) *Composite {
	return &Composite{positionsParsers: parsers}
}

func (c *Composite) ParsePositions(ctx context.Context, wallets []types.Address, block types.BlockRef) ([]Position, error) {
	var positions []Position

	for _, p := range c.positionsParsers {
		pos, err := p.ParsePositions(ctx, wallets, block)
		if err != nil {
			continue
		}
		positions = append(positions, pos...)
	}

	return positions, nil
}
