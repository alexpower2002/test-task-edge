package worker

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"test-task-edge/internal/protocol"
	"test-task-edge/internal/types"
)

const checkInterval = time.Second

type blockPoller interface {
	LatestBlock(ctx context.Context) (types.BlockRef, error)
	BlockByNumber(ctx context.Context, number uint64) (types.BlockRef, error)
}

type positionsParser interface {
	ParsePositions(ctx context.Context, wallets []types.Address, block types.BlockRef) ([]protocol.Position, error)
}

type positionSaver interface {
	SavePositions(ctx context.Context, positions []protocol.Position) error
}

type Worker struct {
	network      string
	poller       blockPoller
	parser       positionsParser
	wallets      []types.Address
	saver        positionSaver
	checkTimeout time.Duration
}

func New(network string, poller blockPoller, parser positionsParser, wallets []types.Address, saver positionSaver) *Worker {
	return &Worker{network: network, poller: poller, parser: parser, wallets: wallets, saver: saver, checkTimeout: checkInterval}
}

func (w *Worker) Run(ctx context.Context) error {
	var last uint64
	for {
		head, err := w.poller.LatestBlock(ctx)
		if err != nil {
			log.Error().Err(err).Str("network", w.network).Msg("failed to read latest block")
		} else if head.Number > last {
			w.processRange(ctx, head, last)
			last = head.Number
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(w.checkTimeout):
		}
	}
}

func (w *Worker) processRange(ctx context.Context, head types.BlockRef, last uint64) {
	from := last + 1
	if last == 0 {
		from = head.Number
	}
	for n := from; n <= head.Number; n++ {
		cur := head
		if n != head.Number {
			var err error
			cur, err = w.poller.BlockByNumber(ctx, n)
			if err != nil {
				log.Error().Err(err).Str("network", w.network).Uint64("block", n).Msg("failed to read block")
				return
			}
		}
		cur.Network = w.network
		w.handleBlock(ctx, cur)
	}
}

func (w *Worker) handleBlock(ctx context.Context, block types.BlockRef) {
	log.Info().Str("network", w.network).Uint64("block", block.Number).Msg("processing block")
	positions, _ := w.parser.ParsePositions(ctx, w.wallets, block)

	for i := range positions {
		positions[i].Network = w.network
	}

	if len(positions) > 0 {
		if err := w.saver.SavePositions(ctx, positions); err != nil {
			log.Error().Err(err).Msg("failed to save positions")
		}
	}
}
