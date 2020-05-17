package rbac

import (
	"github.com/gaia-pipeline/gaia"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"sync"
	"testing"
	"time"
)

var testItem = gaia.RBACEvaluatedPermissions{
	PipelineNamespace: {
		GetAction: map[gaia.RBACPolicyResource]string{
			"*": "allow",
		},
	},
}

var testMap = map[string]cacheItem{
	"item": {
		value:      testItem,
		expiration: time.Now().Add(1 * time.Minute).UnixNano(),
	},
	"item_expired": {
		value:      testItem,
		expiration: time.Now().Add(-1 * time.Minute).UnixNano(),
	},
}

func TestCache_Get_ValidItem_ReturnsValue(t *testing.T) {
	cache := cache{
		expiration: time.Millisecond * 10,
		items:      testMap,
		mu:         sync.Mutex{},
	}

	value, ok := cache.Get("item")

	assert.Check(t, cmp.DeepEqual(value, testMap["item"].value))
	assert.Check(t, cmp.Equal(ok, true))
}

func TestCache_Get_MissingItem_ReturnsEmptyStructAndFalse(t *testing.T) {
	cache := cache{
		expiration: time.Millisecond * 10,
		items:      map[string]cacheItem{},
		mu:         sync.Mutex{},
	}

	value, ok := cache.Get("missing")

	assert.Check(t, cmp.DeepEqual(value, gaia.RBACEvaluatedPermissions{}))
	assert.Check(t, cmp.Equal(ok, false))
}

func TestCache_Get_ExpiredItem_ReturnsEmptyStructAndFalse(t *testing.T) {
	cache := cache{
		expiration: time.Millisecond * 10,
		items:      testMap,
		mu:         sync.Mutex{},
	}

	value, ok := cache.Get("item_expired")

	assert.Check(t, cmp.DeepEqual(value, gaia.RBACEvaluatedPermissions{}))
	assert.Check(t, cmp.Equal(ok, false))
}

func TestCache_Put_Item_ReturnsItem(t *testing.T) {
	cache := cache{
		expiration: time.Minute * 1,
		items:      map[string]cacheItem{},
		mu:         sync.Mutex{},
	}

	value := cache.Put("item", testItem)

	assert.Check(t, cmp.DeepEqual(value, testItem))
}

func TestCache_EvictExpired(t *testing.T) {
	cache := cache{
		expiration: time.Minute * 1,
		items: map[string]cacheItem{
			"expired_01": {
				value:      testItem,
				expiration: time.Now().Add(-2 * time.Minute).UnixNano(),
			},
			"valid_01": {
				value:      testItem,
				expiration: time.Now().Add(2 * time.Minute).UnixNano(),
			},
			"expired_02": {
				value:      testItem,
				expiration: time.Now().Add(-2 * time.Minute).UnixNano(),
			},
		},
		mu: sync.Mutex{},
	}

	cache.EvictExpired()

	_, p1 := cache.items["expired_01"]
	assert.Check(t, cmp.Equal(p1, false))
	_, p2 := cache.items["valid_01"]
	assert.Check(t, cmp.Equal(p2, true))
	_, p3 := cache.items["expired_01"]
	assert.Check(t, cmp.Equal(p3, false))
}

func TestCache_ClearCache(t *testing.T) {
	cache := cache{
		expiration: time.Minute * 1,
		items: map[string]cacheItem{
			"expired_01": {
				value:      testItem,
				expiration: time.Now().Add(-2 * time.Minute).UnixNano(),
			},
			"valid_01": {
				value:      testItem,
				expiration: time.Now().Add(2 * time.Minute).UnixNano(),
			},
		},
		mu: sync.Mutex{},
	}

	cache.Clear()
	assert.Check(t, cmp.Len(cache.items, 0))
}
