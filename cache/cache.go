package cache

// Cache memoizes computed values. It is designed to be efficiently
// reusable by resetting the isSet array (which uses an efficient memset).
// Values for which isSet[i] == false are not defined.
type Cache struct {
	values []interface{}
	isSet  []bool
}

func New(size int) *Cache {
	return &Cache{
		values: make([]interface{}, size),
		isSet:  make([]bool, size),
	}
}

func New2D(size1, size2 int) []*Cache {
	result := make([]*Cache, size1)
	for i := range result {
		result[i] = New(size2)
	}
	return result
}

func (c *Cache) Reset() {
	for i := range c.isSet {
		c.isSet[i] = false
	}
}

func (c *Cache) Set(key uint, value interface{}) {
	if c == nil {
		return
	}

	c.values[key] = value
	c.isSet[key] = true
}

func (c *Cache) Get(key uint) interface{} {
	if c == nil {
		return nil
	}

	return c.values[key]
}

func (c *Cache) IsSet(key uint) bool {
	if c == nil {
		return false
	}

	return c.isSet[key]
}
