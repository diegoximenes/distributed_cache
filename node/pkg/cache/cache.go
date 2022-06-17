package cache

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

const maxTTL int64 = 600

type PutInput struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
	TTL   *int64 `json:"ttl"`
}

type obj struct {
	value   string
	putTime int64
	ttl     *int64
}

type Cache struct {
	lru *lru.Cache
	// uses Mutex instead of RWMutex since Get can remove element due to expiration
	mutex sync.Mutex
}

func New(size int) (*Cache, error) {
	lruCache, err := lru.New(size)
	if err != nil {
		return nil, err
	}

	cache := Cache{
		lru:   lruCache,
		mutex: sync.Mutex{},
	}
	return &cache, nil
}

func (c *Cache) Get(key string) (string, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	objI, exists := c.lru.Get(key)
	if !exists {
		return "", false
	}
	obj := objI.(obj)

	ttl := maxTTL
	if (obj.ttl != nil) && (*obj.ttl < ttl) && (*obj.ttl > 0) {
		ttl = *obj.ttl
	}

	if obj.putTime+ttl < time.Now().Unix() {
		c.lru.Remove(key)
		return "", false
	}

	return obj.value, exists
}

func (c *Cache) Put(input *PutInput) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.lru.Add(input.Key, obj{
		value:   input.Value,
		putTime: time.Now().Unix(),
		ttl:     input.TTL,
	})
}

func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.lru.Remove(key)
}
