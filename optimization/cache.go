package optimization

import (
	"encoding/gob"
	"io"
	"os"
	"sync"

	gzip "github.com/klauspost/pgzip"
)

// Cache memoizes computed values. It is designed to be efficiently
// reusable by resetting the isSet array (which uses an efficient memset).
// Values for which isSet[i] == false are not defined.
type Cache struct {
	mu     sync.RWMutex
	values []GameResult
	isSet  []bool
	nSet   int
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

func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nSet
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.values)
}

func (c *Cache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, ok := range c.isSet {
		if ok {
			c.values[i].Close()
		}
	}

	for i := range c.isSet {
		c.isSet[i] = false
	}

	c.nSet = 0
}

func (c *Cache) Set(key uint, value GameResult) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
	c.isSet[key] = true
	c.nSet++
}

func (c *Cache) Get(key uint) (GameResult, bool) {
	if c == nil {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[key], c.isSet[key]
}

func (c *Cache) IsSet(key uint) bool {
	if c == nil {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isSet[key]
}

type cacheValue struct {
	Key   uint
	Value GameResult
}

func (c *Cache) LoadFromFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzf.Close()

	dec := gob.NewDecoder(gzf)
	for {
		var result cacheValue
		if err := dec.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		c.Set(result.Key, result.Value)
	}

	return nil
}

func (c *Cache) SaveToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	enc := gob.NewEncoder(gzw)
	for key, value := range c.values {
		if c.isSet[key] {
			result := cacheValue{uint(key), value}
			if err := enc.Encode(result); err != nil {
				return err
			}
		}
	}

	return nil
}
