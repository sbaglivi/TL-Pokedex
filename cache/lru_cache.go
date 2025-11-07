package cache

import (
	"container/list"
	"fmt"
	"sync"
)

type Entry struct {
	key   string
	value any
}

type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List
	mu       sync.RWMutex
}

func NewLRU(cap int) *LRUCache {
	if cap <= 0 {
		panic(fmt.Sprintf("LRUCache initialized with capacity %d. Only positive numbers are accepted", cap))
	}

	return &LRUCache{
		capacity: cap,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (cache *LRUCache) Get(key string) (any, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	v, exists := cache.items[key]
	if !exists {
		return nil, false
	}
	cache.order.MoveToFront(v)
	return v.Value.(*Entry).value, true
}

func (cache *LRUCache) Put(key string, value any) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	v, exists := cache.items[key]
	if exists {
		v.Value.(*Entry).value = value
		cache.order.MoveToFront(v)
		return
	}

	value = &Entry{key, value}
	used := len(cache.items)
	if used == cache.capacity {
		evictNode := cache.order.Back()
		evictKey := evictNode.Value.(*Entry).key
		cache.order.Remove(evictNode)
		delete(cache.items, evictKey)
	}

	node := cache.order.PushFront(value)
	cache.items[key] = node
}
