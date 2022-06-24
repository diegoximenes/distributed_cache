package main

import (
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/config"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/httprouter"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/logger"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft"
)

func main() {
	config.Read()

	err := logger.Init()
	if err != nil {
		panic(err)
	}
	defer logger.Logger.Sync()

	raftNode, fsm, raftNodeMetadataClient, err := raft.Set()
	if err != nil {
		panic(err)
	}

	httprouter.Set(raftNode, fsm, raftNodeMetadataClient)
}
