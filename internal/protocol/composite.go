package protocol

import (
	"context"
	"sync"

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
	type result struct {
		positions []Position
	}

	results := make(chan result, len(c.positionsParsers))
	var wg sync.WaitGroup

	for _, p := range c.positionsParsers {
		wg.Add(1)
		go func(parser positionsParser) {
			defer wg.Done()
			pos, err := parser.ParsePositions(ctx, wallets, block)
			if err != nil {
				return
			}
			results <- result{positions: pos}
		}(p)
	}

	wg.Wait()
	close(results)

	var positions []Position
	for r := range results {
		positions = append(positions, r.positions...)
	}

	return positions, nil
}
