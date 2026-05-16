package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"test-task-edge/internal/protocol"
	"test-task-edge/internal/types"
)

type mockSaver struct {
	mu        sync.Mutex
	positions []protocol.Position
}

func (s *mockSaver) SavePositions(_ context.Context, positions []protocol.Position) error {
	s.mu.Lock()
	s.positions = append(s.positions, positions...)
	s.mu.Unlock()
	return nil
}

type stubParserOk struct{}

func (s *stubParserOk) ParsePositions(_ context.Context, _ []types.Address, block types.BlockRef) ([]protocol.Position, error) {
	return []protocol.Position{{
		Protocol:    "mock",
		BlockNumber: block.Number,
	}}, nil
}

type stubPollerFixed struct {
	ref types.BlockRef
}

func (s *stubPollerFixed) LatestBlock(_ context.Context) (types.BlockRef, error) {
	return s.ref, nil
}

func (s *stubPollerFixed) BlockByNumber(_ context.Context, n uint64) (types.BlockRef, error) {
	return types.BlockRef{Number: n}, nil
}

type mockPollerRPCErr struct {
	mu      sync.Mutex
	attempt int
}

func (s *mockPollerRPCErr) LatestBlock(_ context.Context) (types.BlockRef, error) {
	s.mu.Lock()
	s.attempt++
	attempt := s.attempt
	s.mu.Unlock()
	if attempt <= 2 {
		return types.BlockRef{}, errors.New("rpc error")
	}
	return types.BlockRef{Number: 1}, nil
}

func (s *mockPollerRPCErr) BlockByNumber(_ context.Context, n uint64) (types.BlockRef, error) {
	return types.BlockRef{Number: n, Timestamp: 100}, nil
}

type stubPollerHead struct {
	mu   sync.Mutex
	head uint64
}

func (s *stubPollerHead) LatestBlock(_ context.Context) (types.BlockRef, error) {
	s.mu.Lock()
	h := s.head
	s.mu.Unlock()
	return types.BlockRef{Number: h}, nil
}

func (s *stubPollerHead) BlockByNumber(_ context.Context, n uint64) (types.BlockRef, error) {
	return types.BlockRef{Number: n, Timestamp: 100}, nil
}

func TestWorker_savesPositions(t *testing.T) {
	tests := []struct {
		name      string
		poller    blockPoller
		wantBlock uint64
	}{
		{
			name:      "first_block",
			poller:    &stubPollerFixed{ref: types.BlockRef{Number: 1, Timestamp: 100}},
			wantBlock: 1,
		},
		{
			name:      "higher_head",
			poller:    &stubPollerFixed{ref: types.BlockRef{Number: 5}},
			wantBlock: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			saver := &mockSaver{}
			w := New("test", tt.poller, protocol.NewComposite(&stubParserOk{}), nil, saver)

			_ = w.Run(ctx)

			saver.mu.Lock()
			defer saver.mu.Unlock()
			require.Len(t, saver.positions, 1)
			assert.Equal(t, "mock", saver.positions[0].Protocol)
			assert.Equal(t, "test", saver.positions[0].Network)
			assert.Equal(t, tt.wantBlock, saver.positions[0].BlockNumber)
		})
	}
}

func TestWorker_headIncreases(t *testing.T) {
	headStub := &stubPollerHead{head: 1}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go func() {
		time.Sleep(20 * time.Millisecond)
		headStub.mu.Lock()
		headStub.head = 3
		headStub.mu.Unlock()
	}()

	saver := &mockSaver{}
	w := New("test", headStub, protocol.NewComposite(&stubParserOk{}), nil, saver)
	w.checkTimeout = 10 * time.Millisecond

	_ = w.Run(ctx)

	saver.mu.Lock()
	defer saver.mu.Unlock()
	require.Len(t, saver.positions, 3)
	assert.Equal(t, uint64(1), saver.positions[0].BlockNumber)
	assert.Equal(t, uint64(3), saver.positions[2].BlockNumber)
}

func TestWorker_blockByNumberError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	saver := &mockSaver{}
	w := New("test", &mockPollerRPCErr{}, protocol.NewComposite(&stubParserOk{}), nil, saver)

	err := w.Run(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestWorker_contextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	saver := &mockSaver{}
	w := New("test", &stubPollerFixed{ref: types.BlockRef{Number: 1}}, protocol.NewComposite(&stubParserOk{}), nil, saver)

	err := w.Run(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}
