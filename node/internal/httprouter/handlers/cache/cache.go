package cache

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/diegoximenes/distributed_cache/node/internal/logger"
	cacheObj "github.com/diegoximenes/distributed_cache/node/pkg/cache"
)

type GetResponse struct {
	Value interface{} `json:"value"`
}

type APIError struct {
	Error string `json:"error"`
}

func Get(cache *cacheObj.Cache) func(c *gin.Context) {
	return func(c *gin.Context) {
		key := c.Param("key")
		value, exists := cache.Get(key)
		if !exists {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		response := GetResponse{
			Value: value,
		}
		c.JSON(http.StatusOK, response)
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
			c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
		cache.Put(&input)
		c.AbortWithStatus(http.StatusOK)

		logger.Logger.Info(
			"",
			zap.String("handler", "cache.Put"),
			zap.String("input", fmt.Sprintf("%v", input)),
		)
	}
}
