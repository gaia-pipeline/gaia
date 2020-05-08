package cachehelper

import (
	"sync"
	"time"
)

// Cache represents the interface for a simple cache.
type Cache interface {
	Get(key string) (interface{}, bool)
	Put(key string, value interface{}) interface{}
}

type cacheItem struct {
	Value      interface{}
	Expiration int64
}

type cache struct {
	mu         sync.Mutex
	expiration time.Duration
	items      map[string]cacheItem
}

// NewCache creates a new cache. This cache works using expiration based eviction.
func NewCache(expiration time.Duration) Cache {
	return &cache{
		items:      make(map[string]cacheItem),
		expiration: expiration,
		mu:         sync.Mutex{},
	}
}

// Put creates or updates an item. This uses the default expiration time.
func (c *cache) Put(key string, value interface{}) interface{} {
	defer c.mu.Unlock()
	c.mu.Lock()

	c.items[key] = cacheItem{
		Value:      value,
		Expiration: time.Now().Add(c.expiration).UnixNano(),
	}

	item, _ := c.Get(key)
	return item
}

// Get simply gets a item from the cache based on the key.
func (c *cache) Get(key string) (interface{}, bool) {
	item, exists := c.items[key]
	if !exists {
		return nil, false
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}
	return item.Value, exists
}

// EvictExpired evicts any expired items from the cache.
func (c *cache) EvictExpired() {
	now := time.Now().UnixNano()

	defer c.mu.Unlock()
	c.mu.Lock()

	for k, item := range c.items {
		exp := item.Expiration
		if exp > 0 && now > exp {
			delete(c.items, k)
		}
	}
}
