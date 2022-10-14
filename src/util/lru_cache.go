package util

import (
	"github.com/golang/groupcache/lru"
)

type LruCache[K comparable, V any] struct {
	cache		*lru.Cache
	evicted		func(K, V)
}

func NewLruCache[K comparable, V any](max_size int, evicted func(K, V)) *LruCache[K, V] {
	lru_cache := &LruCache[K, V] {
		cache:		lru.New(max_size),
		evicted: 	evicted,
	}
	lru_cache.cache.OnEvicted = lru_cache.do_evicted
	return lru_cache
}

func (p *LruCache[K, V]) Set(key K, value V) {
	p.cache.Add(key, value)
}

func (p *LruCache[K, V]) Get(key K) (value V, ok bool) {
	if v, ok := p.cache.Get(key); ok {
		return v.(V), true
	}
	return
}

func (p *LruCache[K, V]) Remove(key K) {
	p.cache.Remove(key)
}

func (p *LruCache[K, V]) RemoveOldest() {
	p.cache.RemoveOldest()
}

func (p *LruCache[K, V]) Len() int {
	return p.cache.Len()
}

func (p *LruCache[K, V]) do_evicted(k lru.Key, v interface {}) {
	key := k.(K)
	value := v.(V)
	p.evicted(key, value)
}