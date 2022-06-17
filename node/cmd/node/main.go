package main

import (
	"github.com/gin-gonic/gin"

	"github.com/diegoximenes/distributed_cache/node/internal/config"
	cacheHandler "github.com/diegoximenes/distributed_cache/node/internal/handlers/cache"
	"github.com/diegoximenes/distributed_cache/node/internal/handlers/heartbeat"
	cache "github.com/diegoximenes/distributed_cache/node/pkg/cache"
)

func main() {
	config.Read()

	cache, err := cache.New(config.Config.CacheSize)
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.GET("/cache/:key", cacheHandler.Get(cache))
	router.DELETE("/cache/:key", cacheHandler.Delete(cache))
	router.PUT("/cache", cacheHandler.Put(cache))
	router.GET("/heartbeat", heartbeat.Heartbeat)
	router.Run()
}
