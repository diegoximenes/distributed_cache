package main

import (
	"github.com/gin-gonic/gin"

	cacheHandler "github.com/diegoximenes/distributed_cache/node/internal/handlers/cache"
	"github.com/diegoximenes/distributed_cache/node/internal/handlers/heartbeat"
	cache "github.com/diegoximenes/distributed_cache/node/pkg/cache"
)

func main() {
	cache := cache.New()

	router := gin.Default()
	router.GET("/cache/:key", cacheHandler.Get(cache))
	router.DELETE("/cache/:key", cacheHandler.Delete(cache))
	router.PUT("/cache", cacheHandler.Put(cache))
	router.GET("/heartbeat", heartbeat.Heartbeat)
	router.Run()
}
