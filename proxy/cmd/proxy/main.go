package main

import (
	"github.com/gin-gonic/gin"

	"github.com/diegoximenes/distributed_cache/proxy/internal/config"
	"github.com/diegoximenes/distributed_cache/proxy/internal/handlers/cache"
	"github.com/diegoximenes/distributed_cache/proxy/internal/handlers/heartbeat"
	"github.com/diegoximenes/distributed_cache/proxy/internal/util/logger"
	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/nodesmetadata"
)

func main() {
	config.Read()

	err := logger.Init()
	if err != nil {
		panic(err)
	}

	nodesMetadataClient, err := nodesmetadata.New()
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.GET("/heartbeat", heartbeat.Heartbeat)
	router.GET("/cache/:key", cache.Get(nodesMetadataClient))
	router.PUT("/cache", cache.Put(nodesMetadataClient))
	router.Run()
}
