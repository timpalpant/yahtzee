package optimization

// Cache memoizes computed values used within a Turn.
type Cache struct {
	values []GameResult
	isSet  []bool
}

func NewCache(size int) *Cache {
	return &Cache{
		values: make([]GameResult, size),
		isSet:  make([]bool, size),
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

func (c *Cache) Reset() {
	for i, isSet := range c.isSet {
		if isSet {
			c.values[i].Close()
		}
	}

	for i := range c.isSet {
		c.isSet[i] = false
	}
}

func (c *Cache) Set(key uint, value GameResult) {
	if c == nil {
		return
	}

	c.values[key] = value
	c.isSet[key] = true
}

func (c *Cache) Get(key uint) (GameResult, bool) {
	if c == nil {
		return nil, false
	}

	ok := c.isSet[key]
	value := c.values[key]
	return value, ok
}
