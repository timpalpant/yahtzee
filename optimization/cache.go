package optimization

// Cache memoizes computed values used within a Turn.
type Cache struct {
	values      []GameResult
	isSet       []bool
	denseValues []GameResult
}

func NewCache(size int) *Cache {
	return &Cache{
		values:      make([]GameResult, size),
		isSet:       make([]bool, size),
		denseValues: make([]GameResult, 0),
	}
}

func New2DCache(size1, size2 int) []*Cache {
	result := make([]*Cache, size1)
	for i := range result {
		result[i] = NewCache(size2)
	}
	return result
}

func (c *Cache) Len() int {
	return len(c.values)
}

func (c *Cache) Reset() []GameResult {
	for i := range c.isSet {
		c.isSet[i] = false
	}

	result := c.denseValues
	c.denseValues = c.denseValues[:0]
	return result
}

func (c *Cache) Set(key uint, value GameResult) {
	if c == nil {
		return
	}

	c.values[key] = value
	c.isSet[key] = true
	c.denseValues = append(c.denseValues, value)
}

func (c *Cache) Get(key uint) (GameResult, bool) {
	if c == nil {
		return nil, false
	}

	ok := c.isSet[key]
	value := c.values[key]
	return value, ok
}
