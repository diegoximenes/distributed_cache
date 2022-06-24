package httprouter

import (
	cacheHandler "github.com/diegoximenes/distributed_cache/node/internal/httprouter/handlers/cache"
	"github.com/diegoximenes/distributed_cache/node/internal/httprouter/handlers/heartbeat"
	"github.com/diegoximenes/distributed_cache/node/pkg/cache"
	"github.com/gin-gonic/gin"
)

func Set(cache *cache.Cache) {
	router := gin.Default()
	router.GET("/cache/:key", cacheHandler.Get(cache))
	router.DELETE("/cache/:key", cacheHandler.Delete(cache))
	router.PUT("/cache", cacheHandler.Put(cache))
	router.GET("/heartbeat", heartbeat.Heartbeat)
	router.Run()
}
