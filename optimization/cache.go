package optimization

// Cache memoizes computed values and is thread-safe.
type Cache struct {
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
	return len(c.values)
}

func (c *Cache) Reset() {
	for key, value := range c.values {
		value.Close()
		delete(c.values, key)
	}
}

func (c *Cache) Set(key uint, value GameResult) {
	if c == nil {
		return
	}

	c.values[key] = value
}

func (c *Cache) Get(key uint) (GameResult, bool) {
	if c == nil {
		return nil, false
	}

	value, ok := c.values[key]
	return value, ok
}
