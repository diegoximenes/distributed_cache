package main

import (
	"github.com/diegoximenes/distributed_cache/proxy/internal/config"
	"github.com/diegoximenes/distributed_cache/proxy/internal/httprouter"
	"github.com/diegoximenes/distributed_cache/proxy/internal/keypartition"
	"github.com/diegoximenes/distributed_cache/proxy/internal/logger"
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

	httprouter.Set(nodeClient, nodesMetadataClient, keyPartitionStrategy)
}
