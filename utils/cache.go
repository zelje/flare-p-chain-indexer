package utils

import (
	"container/list"
	"sync"
)

type Cache[K comparable, V any] interface {
	Add(K, V)
	Get(K) (V, bool)
}

// Map object cache
type cache[K comparable, V any] struct {
	sync.RWMutex

	cacheMap map[K]V
	keys     *list.List
	maxSize  int
}

func NewCache[K comparable, V any](maxSize int) Cache[K, V] {
	return &cache[K, V]{
		cacheMap: make(map[K]V),
		keys:     list.New(),
		maxSize:  maxSize,
	}
}

func (c *cache[K, V]) Add(k K, v V) {
	c.Lock()
	if _, ok := c.cacheMap[k]; ok {
		c.cacheMap[k] = v
	} else {
		c.cacheMap[k] = v
		c.keys.PushBack(k)
		if c.keys.Len() > c.maxSize {
			e := c.keys.Front()
			c.keys.Remove(e)
			delete(c.cacheMap, e.Value.(K))
		}
	}
	c.Unlock()
}

func (c *cache[K, V]) Get(k K) (V, bool) {
	v, ok := c.cacheMap[k]
	return v, ok
}
