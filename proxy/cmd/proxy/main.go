package main

import (
	"github.com/gin-gonic/gin"

	"github.com/diegoximenes/distributed_cache/proxy/internal/config"
	"github.com/diegoximenes/distributed_cache/proxy/internal/handlers/cache"
	"github.com/diegoximenes/distributed_cache/proxy/internal/handlers/heartbeat"
	"github.com/diegoximenes/distributed_cache/proxy/internal/keypartition"
	"github.com/diegoximenes/distributed_cache/proxy/internal/util/logger"
	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/node"
	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/nodesmetadata"
)

func main() {
	config.Read()

	err := logger.Init()
	if err != nil {
		panic(err)
	}
	defer logger.Logger.Sync()

	keyPartitionStrategy := keypartition.New()

	nodesMetadataClient := nodesmetadata.New(keyPartitionStrategy)

	nodeClient := node.New()

	router := gin.Default()
	router.GET("/heartbeat", heartbeat.Heartbeat)
	router.GET("/cache/:key", cache.Get(nodeClient, nodesMetadataClient, keyPartitionStrategy))
	router.PUT("/cache", cache.Put(nodeClient, nodesMetadataClient, keyPartitionStrategy))
	router.Run()
}
