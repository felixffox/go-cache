package solution

import (
	"sync"
	"time"
)

const defaultLifetime = 5 * time.Minute

type CacheOption func(*Cache)

type Item struct {
	value      interface{}
	expiration int64
}

type cleaner struct {
	key       uint64
	timestamp int64
}

type Cache struct {
	lifetime   time.Duration
	storage    map[uint64]*Item
	rwMutex    sync.RWMutex
	deathQueue chan cleaner
}

func New(opts ...CacheOption) *Cache {
	cache := &Cache{
		lifetime:   defaultLifetime,
		storage:    make(map[uint64]*Item),
		rwMutex:    sync.RWMutex{},
		deathQueue: make(chan cleaner, 100),
	}

	for _, opt := range opts {
		opt(cache)
	}

	go cache.startCleanupDaemon()

	return cache
}

func (c *Cache) Set(key uint64, value interface{}) {
	expiration := time.Now().Add(c.lifetime).UnixNano()

	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.storage[key] = &Item{
		value:      value,
		expiration: expiration,
	}

	c.deathQueue <- cleaner{key: key, timestamp: expiration}
}

func (c *Cache) Get(key uint64) (interface{}, bool) {
	c.rwMutex.RLock()
	obj, exists := c.storage[key]
	c.rwMutex.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().UnixNano() > obj.expiration {
		c.rwMutex.Lock()
		defer c.rwMutex.Unlock()
		delete(c.storage, key)
		return nil, false
	}

	return obj.value, true
}

func (c *Cache) startCleanupDaemon() {
	for cleanerObj := range c.deathQueue {
		timeUntilExpiration := time.Until(time.Unix(0, cleanerObj.timestamp))

		if timeUntilExpiration > 0 {
			time.Sleep(timeUntilExpiration)
		}

		c.rwMutex.Lock()
		obj, exists := c.storage[cleanerObj.key]
		if exists && obj.expiration == cleanerObj.timestamp {
			delete(c.storage, cleanerObj.key)
		}
		c.rwMutex.Unlock()
	}
}

func WithCustomLifetime(ttl time.Duration) CacheOption {
	return func(c *Cache) {
		c.lifetime = ttl
	}
}
