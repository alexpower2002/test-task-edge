package protocol

import (
	"context"
	"sync"

	"test-task-edge/internal/types"
)

type positionsParser interface {
	ParsePositions(ctx context.Context, wallets []types.Address, block types.BlockRef) ([]Position, error)
}

type SubParser struct {
	Parser  positionsParser
	Wallets []types.Address
}

type Composite struct {
	parsers []SubParser
}

func NewComposite(parsers ...SubParser) *Composite {
	return &Composite{parsers: parsers}
}

func (c *Composite) ParsePositions(ctx context.Context, _ []types.Address, block types.BlockRef) ([]Position, error) {
	type result struct {
		positions []Position
	}

	results := make(chan result, len(c.parsers))
	var wg sync.WaitGroup

	for _, sp := range c.parsers {
		wg.Add(1)
		go func(sp SubParser) {
			defer wg.Done()
			pos, err := sp.Parser.ParsePositions(ctx, sp.Wallets, block)
			if err != nil {
				return
			}
			results <- result{positions: pos}
		}(sp)
	}

	wg.Wait()
	close(results)

	var positions []Position
	for r := range results {
		positions = append(positions, r.positions...)
	}

	return positions, nil
}
