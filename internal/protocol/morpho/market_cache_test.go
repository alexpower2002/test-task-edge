package morpho

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"test-task-edge/internal/types"
)

func TestMarketCache_miss(t *testing.T) {
	c := newMarketCache()
	_, ok := c.get()
	assert.False(t, ok)
}

func TestMarketCache_set_get(t *testing.T) {
	c := newMarketCache()
	ids := []types.Bytes32{{1}, {2}}
	c.set(ids)
	got, ok := c.get()
	assert.True(t, ok)
	assert.Equal(t, ids, got)
}

func TestMarketCache_get_twice(t *testing.T) {
	c := newMarketCache()
	ids := []types.Bytes32{{1}}
	c.set(ids)
	got1, ok1 := c.get()
	assert.True(t, ok1)
	assert.Equal(t, ids, got1)
	got2, ok2 := c.get()
	assert.True(t, ok2)
	assert.Equal(t, ids, got2)
}

func TestMarketCache_overwrite(t *testing.T) {
	c := newMarketCache()
	c.set([]types.Bytes32{{1}})
	c.set([]types.Bytes32{{2}})
	got, ok := c.get()
	assert.True(t, ok)
	assert.Equal(t, []types.Bytes32{{2}}, got)
}

func TestMarketCache_empty_slice(t *testing.T) {
	c := newMarketCache()
	c.set([]types.Bytes32{})
	got, ok := c.get()
	assert.True(t, ok)
	assert.Empty(t, got)
}
