package cachehelper

import (
	"sync"
	"time"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type Cache struct {
	mu         sync.Mutex
	expiration time.Duration
	items      map[string]CacheItem
}

func NewCache(expiration time.Duration) *Cache {
	return &Cache{
		items:      make(map[string]CacheItem),
		expiration: expiration,
		mu:         sync.Mutex{},
	}
}

func (c *Cache) Put(key string, value interface{}) interface{} {
	defer c.mu.Unlock()
	c.mu.Lock()

	c.items[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Add(c.expiration).UnixNano(),
	}

	item, _ := c.Get(key)
	return item
}

func (c *Cache) Get(key string) (interface{}, bool) {
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

func (c *Cache) EvictExpired() {
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
