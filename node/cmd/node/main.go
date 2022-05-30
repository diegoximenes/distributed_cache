package main

import (
	"github.com/gin-gonic/gin"

	cacheHandler "github.com/diegoximenes/distributed_key_value_cache/node/internal/handlers/cache"
	heartbeatHandler "github.com/diegoximenes/distributed_key_value_cache/node/internal/handlers/heartbeat"
	cache "github.com/diegoximenes/distributed_key_value_cache/node/pkg/cache"
)

func main() {
	cache := cache.New()

	router := gin.Default()
	router.GET("/cache/:key", cacheHandler.Get(cache))
	router.PUT("/cache", cacheHandler.Put(cache))
	router.GET("/heartbeat", heartbeatHandler.Heartbeat)
	router.Run()
}
