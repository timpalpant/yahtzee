package yahtzee

import (
	"bufio"
	"encoding/json"
	"os"
)

// Cache memoizes computed values. It is designed to be efficiently
// reusable by resetting the isSet array (which uses an efficient memset).
// Values for which isSet[i] == false are not defined.
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

type cacheValue struct {
	Key   uint
	Value GameResult
}

func (c *Cache) SaveToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	defer buf.Flush()

	enc := json.NewEncoder(buf)
	for key, isSet := range c.isSet {
		if isSet {
			value := c.values[key]
			result := cacheValue{uint(key), value}
			if err := enc.Encode(&result); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Cache) Reset() {
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

func (c *Cache) Get(key uint) GameResult {
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
