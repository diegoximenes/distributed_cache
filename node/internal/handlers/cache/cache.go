package cache

import (
	"net/http"

	"github.com/gin-gonic/gin"

	cacheObj "github.com/diegoximenes/distributed_key_value_cache/node/pkg/cache"
)

type GetResponse struct {
	Value string `json:"value"`
}

func Get(cache *cacheObj.Cache) func(c *gin.Context) {
	return func(c *gin.Context) {
		key := c.Param("key")
		value, exists := cache.Get(key)
		if !exists {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			response := GetResponse{
				Value: value,
			}
			c.JSON(http.StatusOK, response)
		}
	}
}

type PutInput struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

func Put(cache *cacheObj.Cache) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input PutInput
		c.BindJSON(&input)
		cache.Put(input.Key, input.Value)
		c.AbortWithStatus(http.StatusOK)
	}
}
