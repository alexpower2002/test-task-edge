package morpho

import (
	"sync"

	"test-task-edge/internal/types"
)

type marketCache struct {
	mu     sync.RWMutex
	cached []types.Bytes32
}

func newMarketCache() *marketCache {
	return &marketCache{}
}

func (c *marketCache) get() ([]types.Bytes32, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cached != nil {
		return c.cached, true
	}
	return nil, false
}

func (c *marketCache) set(ids []types.Bytes32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cached = ids
}
