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

// LRUResultCache is a bounded in-memory LRU cache implementation.
type LRUResultCache struct {
	mu       sync.Mutex
	capacity int
	order    *list.List
	entries  map[string]*list.Element
}

type cacheEntry struct {
	key    string
	result Result
}

// NewLRUResultCache creates a bounded in-memory LRU parse result cache.
//
// Returns nil when capacity is <= 0.
func NewLRUResultCache(capacity int) *LRUResultCache {
	if capacity <= 0 {
		return nil
	}
	return newLRUResultCache(capacity)
}

func newLRUResultCache(capacity int) *LRUResultCache {
	if capacity <= 0 {
		return nil
	}
	return &LRUResultCache{
		capacity: capacity,
		order:    list.New(),
		entries:  make(map[string]*list.Element, capacity),
	}
}

// Get returns a cached result for key and updates recency on hit.
func (c *LRUResultCache) Get(key string) (Result, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.entries[key]
	if !ok {
		return Result{}, false
	}
	c.order.MoveToFront(elem)
	return elem.Value.(cacheEntry).result, true
}

// Set stores a result for key and evicts the least-recently-used entry when full.
func (c *LRUResultCache) Set(key string, result Result) {
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

// MemoryResultCache is a simple unbounded in-memory cache implementation.
type MemoryResultCache struct {
	mu      sync.RWMutex
	entries map[string]Result
}

// NewMemoryResultCache creates an unbounded in-memory cache.
func NewMemoryResultCache() *MemoryResultCache {
	return &MemoryResultCache{
		entries: make(map[string]Result),
	}
}

// Get returns a cached result for key.
func (c *MemoryResultCache) Get(key string) (Result, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, ok := c.entries[key]
	return result, ok
}

// Set stores a cached result for key.
func (c *MemoryResultCache) Set(key string, result Result) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = result
}
