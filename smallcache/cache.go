package smallcache

import (
	"SmallCache/smallcache/lru"
	"sync"
)

type cache struct {
	mu         sync.Mutex	// 通过互斥锁，实现并发时的安全
	lru        *lru.Cache	// LRU实例化
	cacheBytes int64		// 缓存大小
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// lazy load
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
