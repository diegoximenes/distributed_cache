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

	keyPartitionStrategy := keypartition.New()

	nodesMetadataClient, err := nodesmetadata.New(keyPartitionStrategy)
	if err != nil {
		panic(err)
	}

	nodeClient := node.New()

	router := gin.Default()
	router.GET("/heartbeat", heartbeat.Heartbeat)
	router.GET("/cache/:key", cache.Get(nodeClient, nodesMetadataClient, keyPartitionStrategy))
	router.PUT("/cache", cache.Put(nodeClient, nodesMetadataClient, keyPartitionStrategy))
	router.Run()
}
