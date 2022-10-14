package hbservice_test

import (
	"testing"
	"hbservice/src/util"
)

func TestLruCache(t *testing.T) {
	fevicted := func(key string, value int) {
		t.Logf("on_evicted|key:%s, value:%d\n", key, value)
	}
	cache := util.NewLruCache[string, int](3, fevicted)
	cache.Set("haobi1", 1)
	cache.Set("haobi2", 2)
	cache.Set("haobi3", 3)
	check_lru_get(t, cache, "haobi1", 1)
	check_lru_get(t, cache, "haobi2", 2)
	check_lru_get(t, cache, "haobi3", 3)

	cache.Set("haobi4", 4)
	check_lru_not_exists(t, cache, "haobi1")

	if cache.Len() != 3 {
		t.Fatalf("cache.Len() != 3")
	}
	check_lru_get(t, cache, "haobi2", 2)
	check_lru_get(t, cache, "haobi3", 3)
	check_lru_get(t, cache, "haobi4", 4)

	cache.Remove("haobi3")
	check_lru_not_exists(t, cache, "haobi3")

	cache.RemoveOldest()
	check_lru_not_exists(t, cache, "haobi2")
}

func check_lru_get(t *testing.T, cache *util.LruCache[string, int], key string, value int) {
	if v, ok := cache.Get(key); ok {
		if v != value {
			t.Fatalf("check_lru_get|not_match, key:%s, value:%d", key, value)
		}
	} else {
		t.Fatalf("check_lru_get|not_exist, key:%s, value:%d", key, value)
	}
}

func check_lru_not_exists(t *testing.T, cache *util.LruCache[string, int], key string) {
	if _, ok := cache.Get(key); ok {
		t.Fatalf("check_lru_not_exists|exists key:%s", key)
	}
}