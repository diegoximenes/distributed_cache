package httprouter

import (
	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/config"
	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/handlers/node"
	raftHandler "github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/handlers/raft"
	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/raft/fsm"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
)

func Set(raftNode *raft.Raft, fsm *fsm.FSM) {
	router := gin.Default()
	router.GET("/node", node.Get(raftNode, fsm))
	router.PUT("/node", node.Put(raftNode))
	router.DELETE("/node/:id", node.Delete(raftNode))
	router.PUT("/raft/join", raftHandler.Join(raftNode))
	router.Run(config.Config.HTTPAddress)
}
