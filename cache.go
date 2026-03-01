package ddgo

import (
	"container/list"
	"sync"
)

type resultCache struct {
	mu       sync.Mutex
	capacity int
	order    *list.List
	entries  map[string]*list.Element
}

type cacheEntry struct {
	key    string
	result Result
}

func newResultCache(capacity int) *resultCache {
	if capacity <= 0 {
		return nil
	}
	return &resultCache{
		capacity: capacity,
		order:    list.New(),
		entries:  make(map[string]*list.Element, capacity),
	}
}

func (c *resultCache) get(key string) (Result, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.entries[key]
	if !ok {
		return Result{}, false
	}
	c.order.MoveToFront(elem)
	return elem.Value.(cacheEntry).result, true
}

func (c *resultCache) set(key string, result Result) {
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
