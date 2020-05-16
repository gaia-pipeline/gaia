package rbac

import (
	"sync"
	"time"

	"github.com/gaia-pipeline/gaia"
)

// Cache represents the interface for a simple cache.
type Cache interface {
	Get(key string) (gaia.RBACEvaluatedPermissions, bool)
	Put(key string, value gaia.RBACEvaluatedPermissions) gaia.RBACEvaluatedPermissions
	EvictExpired()
	Clear()
}

type cacheItem struct {
	value      gaia.RBACEvaluatedPermissions
	expiration int64
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
func (c *cache) Put(key string, value gaia.RBACEvaluatedPermissions) gaia.RBACEvaluatedPermissions {
	defer c.mu.Unlock()
	c.mu.Lock()

	c.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(c.expiration).UnixNano(),
	}

	item, _ := c.Get(key)
	return item
}

// Get simply gets a item from the cache based on the key.
func (c *cache) Get(key string) (gaia.RBACEvaluatedPermissions, bool) {
	item, exists := c.items[key]
	if !exists {
		return gaia.RBACEvaluatedPermissions{}, false
	}
	if item.expiration > 0 {
		if time.Now().UnixNano() > item.expiration {
			return gaia.RBACEvaluatedPermissions{}, false
		}
	}
	return item.value, exists
}

// EvictExpired evicts any expired items from the cache.
func (c *cache) EvictExpired() {
	now := time.Now().UnixNano()

	defer c.mu.Unlock()
	c.mu.Lock()

	for k, item := range c.items {
		exp := item.expiration
		if exp > 0 && now > exp {
			delete(c.items, k)
		}
	}
}

// Clear and invalidate the whole cache
func (c *cache) Clear() {
	defer c.mu.Unlock()
	c.mu.Lock()

	c.items = make(map[string]cacheItem)
}
