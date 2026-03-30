package cache

import (
	"context"
	"sync"
	"time"

	"github.com/voidroute/shortener/internal/domain"
)

type InMemoryCache struct {
	mu      sync.RWMutex
	store   map[string]Item
	ttl     time.Duration
	cleaner *Cleaner
}

type Item struct {
	link      *domain.Link
	expiresAt time.Time
}

func NewInMemoryCache(ttl time.Duration, cleanupInterval time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		store: make(map[string]Item),
		ttl:   ttl,
	}

	cleaner := NewCleaner(cleanupInterval, cache.deleteExpired)
	cleaner.Start()
	cache.cleaner = cleaner

	return cache
}

func (c *InMemoryCache) Set(_ context.Context, code string, link *domain.Link) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[code] = Item{
		link:      link,
		expiresAt: time.Now().UTC().Add(c.ttl),
	}
	return nil
}

func (c *InMemoryCache) Get(_ context.Context, code string) (*domain.Link, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.store[code]
	if !ok {
		return nil, ErrCacheMiss
	}

	if time.Now().UTC().After(item.expiresAt) {
		return nil, ErrCacheExpired
	}

	return item.link, nil
}

func (c *InMemoryCache) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UTC()
	for key, item := range c.store {
		if now.After(item.expiresAt) {
			delete(c.store, key)
		}
	}
}

func (c *InMemoryCache) Ping(_ context.Context) error {
	return nil
}

func (c *InMemoryCache) Close() error {
	if c.cleaner != nil {
		c.cleaner.Stop()
	}
	return nil
}
