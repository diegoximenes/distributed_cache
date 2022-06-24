package httprouter

import (
	"github.com/diegoximenes/distributed_cache/proxy/internal/httprouter/handlers/cache"
	"github.com/diegoximenes/distributed_cache/proxy/internal/httprouter/handlers/heartbeat"
	"github.com/diegoximenes/distributed_cache/proxy/internal/keypartition"
	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/node"
	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/nodesmetadata"
	"github.com/gin-gonic/gin"
)

func Set(
	nodeClient *node.NodeClient,
	nodesMetadataClient *nodesmetadata.NodesMetadataClient,
	keyPartitionStrategy keypartition.KeyPartitionStrategy,
) {
	router := gin.Default()
	router.GET("/heartbeat", heartbeat.Heartbeat)
	router.GET("/cache/:key", cache.Get(nodeClient, nodesMetadataClient, keyPartitionStrategy))
	router.PUT("/cache", cache.Put(nodeClient, nodesMetadataClient, keyPartitionStrategy))
	router.Run()
}
