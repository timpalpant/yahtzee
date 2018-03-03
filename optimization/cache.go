package optimization

import (
	"encoding/gob"
	"io"
	"os"
	"sync"

	"github.com/golang/glog"

	gzip "github.com/klauspost/pgzip"
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

func (c *Cache) Remove(key uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.values, key)
}

func (c *Cache) RemoveAll(keys []uint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		delete(c.values, key)
	}
}

// Attempts to deduplicate the stored values.
// Keys with identical value content will be updated to point
// to a shared address.
func (c *Cache) Compress() {
	c.mu.Lock()
	defer c.mu.Unlock()

	glog.V(1).Infof("Attempting to deduplicate %v values", len(c.values))
	byHash := make(map[string]GameResult)
	nDuplicates := 0
	for key, value := range c.values {
		h := value.HashCode()
		if other, ok := byHash[h]; ok {
			nDuplicates++
			c.values[key] = other
			value.Close()
		} else {
			byHash[h] = value
		}
	}

	glog.V(1).Infof("Deduplicated %v values", nDuplicates)
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
		result := cacheValue{uint(key), value}
		if err := enc.Encode(result); err != nil {
			return err
		}
	}

	return nil
}
