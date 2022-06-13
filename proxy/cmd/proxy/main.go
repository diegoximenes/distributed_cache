package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/diegoximenes/distributed_cache/proxy/internal/config"
	"github.com/diegoximenes/distributed_cache/proxy/internal/handlers/cache"
	"github.com/diegoximenes/distributed_cache/proxy/internal/handlers/heartbeat"
	"github.com/diegoximenes/distributed_cache/proxy/internal/util/logger"
	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/nodemetadata"
)

func main() {
	config.Read()

	err := logger.Init()
	if err != nil {
		panic(fmt.Sprintf("Error when setting logger: %v", err))
	}

	nodeMetadataClient, err := nodemetadata.New()
	if err != nil {
		panic(fmt.Sprintf("Error when setting nodeMetadataClient: %v", err))
	}

	router := gin.Default()
	router.GET("/heartbeat", heartbeat.Heartbeat)
	router.GET("/cache/:key", cache.Get(nodeMetadataClient))
	router.PUT("/cache", cache.Put(nodeMetadataClient))
	router.Run()
}
