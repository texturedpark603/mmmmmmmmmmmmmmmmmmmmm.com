package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     any
	expiresAt time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]entry
	ttl     time.Duration
}

func New(ttl time.Duration) *Cache {
	c := &Cache{
		entries: make(map[string]entry),
		ttl:     ttl,
	}
	go c.sweep()
	return c
}

func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = entry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

func (c *Cache) sweep() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, e := range c.entries {
			if now.After(e.expiresAt) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}
