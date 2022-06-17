package cache

import (
	"net/http"

	"github.com/gin-gonic/gin"

	cacheObj "github.com/diegoximenes/distributed_cache/node/pkg/cache"
)

type GetResponse struct {
	Value interface{} `json:"value"`
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

func Delete(cache *cacheObj.Cache) func(c *gin.Context) {
	return func(c *gin.Context) {
		key := c.Param("key")
		cache.Delete(key)
		c.AbortWithStatus(http.StatusOK)
	}
}

func Put(cache *cacheObj.Cache) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input cacheObj.PutInput
		err := c.BindJSON(&input)
		if err != nil {
			c.Abort()
		}
		if !c.IsAborted() {
			cache.Put(&input)
			c.AbortWithStatus(http.StatusOK)
		}
	}
}
