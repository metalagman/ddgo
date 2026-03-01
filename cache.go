package ddgo

import (
	"container/list"
	"sync"
)

// ResultCache is a pluggable cache used by Detector.Parse.
//
// Implementations must be safe for concurrent use.
type ResultCache interface {
	Get(key string) (Result, bool)
	Set(key string, result Result)
}

type lruResultCache struct {
	mu       sync.Mutex
	capacity int
	order    *list.List
	entries  map[string]*list.Element
}

type cacheEntry struct {
	key    string
	result Result
}

func newResultCache(capacity int) ResultCache {
	cache := newLRUResultCache(capacity)
	if cache == nil {
		return nil
	}
	return cache
}

func newLRUResultCache(capacity int) *lruResultCache {
	if capacity <= 0 {
		return nil
	}
	return &lruResultCache{
		capacity: capacity,
		order:    list.New(),
		entries:  make(map[string]*list.Element, capacity),
	}
}

func (c *lruResultCache) Get(key string) (Result, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.entries[key]
	if !ok {
		return Result{}, false
	}
	c.order.MoveToFront(elem)
	return elem.Value.(cacheEntry).result, true
}

func (c *lruResultCache) Set(key string, result Result) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.entries[key]; ok {
		entry := elem.Value.(cacheEntry)
		entry.result = result
		elem.Value = entry
		c.order.MoveToFront(elem)
		return
	}

	elem := c.order.PushFront(cacheEntry{key: key, result: result})
	c.entries[key] = elem
	if c.order.Len() <= c.capacity {
		return
	}
	tail := c.order.Back()
	if tail == nil {
		return
	}
	entry := tail.Value.(cacheEntry)
	delete(c.entries, entry.key)
	c.order.Remove(tail)
}

// NewMemoryResultCache returns a simple unbounded in-memory cache implementation.
//
// Use this with WithResultCache when you want custom cache behavior without LRU
// eviction logic.
func NewMemoryResultCache() ResultCache {
	return &memoryResultCache{
		entries: make(map[string]Result),
	}
}

type memoryResultCache struct {
	mu      sync.RWMutex
	entries map[string]Result
}

func (c *memoryResultCache) Get(key string) (Result, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, ok := c.entries[key]
	return result, ok
}

func (c *memoryResultCache) Set(key string, result Result) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = result
}
