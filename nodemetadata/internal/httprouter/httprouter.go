package httprouter

import (
	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/config"
	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/httprouter/handlers/node"
	raftHandler "github.com/diegoximenes/distributed_cache/nodemetadata/internal/httprouter/handlers/raft"
	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/fsm"
	raftMetadata "github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/metadata"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
)

func Set(raftNode *raft.Raft, fsm *fsm.FSM, raftNodeMetadataClient *raftMetadata.RaftNodeMetadataClient) {
	router := gin.Default()
	router.GET("/node", node.Get(raftNode, fsm, raftNodeMetadataClient))
	router.PUT("/node", node.Put(raftNode, raftNodeMetadataClient))
	router.DELETE("/node/:id", node.Delete(raftNode, raftNodeMetadataClient))
	router.PUT("/raft/join", raftHandler.Join(raftNode))
	router.Run(config.Config.ApplicationAddress)
}
