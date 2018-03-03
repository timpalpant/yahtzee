package optimization

import (
	"sync"
)

// Cache memoizes computed values and is thread-safe.
type Cache struct {
	mu     sync.RWMutex
	values map[uint]GameResult
}

func NewCache() *Cache {
	return &Cache{
		values: make(map[uint]GameResult),
	}
}

func New2DCache(size1, size2 int) []*Cache {
	result := make([]*Cache, size1)
	for i := range result {
		result[i] = NewCache()
	}
	return result
}

func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.values)
}

func (c *Cache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, value := range c.values {
		value.Close()
		delete(c.values, key)
	}
}

func (c *Cache) Set(key uint, value GameResult) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

func (c *Cache) Get(key uint) (GameResult, bool) {
	if c == nil {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.values[key]
	return value, ok
}
