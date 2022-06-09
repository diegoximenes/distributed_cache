package cache

import "time"

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
	hashTable map[string]obj
}

func New() *Cache {
	return &Cache{
		hashTable: make(map[string]obj),
	}
}

func (c *Cache) Get(key string) (string, bool) {
	obj, exists := c.hashTable[key]

	if !exists {
		return "", false
	}

	ttl := maxTTL
	if (obj.ttl != nil) && (*obj.ttl < ttl) && (*obj.ttl > 0) {
		ttl = *obj.ttl
	}

	if obj.putTime+ttl < time.Now().Unix() {
		delete(c.hashTable, key)
		return "", false
	}

	return obj.value, exists
}

func (c *Cache) Put(input *PutInput) {
	c.hashTable[input.Key] = obj{
		value:   input.Value,
		putTime: time.Now().Unix(),
		ttl:     input.TTL,
	}
}

func (c *Cache) Delete(key string) {
	delete(c.hashTable, key)
}
