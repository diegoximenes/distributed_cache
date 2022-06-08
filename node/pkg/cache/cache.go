package cache

type Cache struct {
	hashTable map[string]string
}

func New() *Cache {
	return &Cache{
		hashTable: make(map[string]string),
	}
}

func (c *Cache) Get(key string) (string, bool) {
	value, exists := c.hashTable[key]
	return value, exists
}

func (c *Cache) Put(key string, value string) {
	c.hashTable[key] = value
}

func (c *Cache) Delete(key string) {
	delete(c.hashTable, key)
}
