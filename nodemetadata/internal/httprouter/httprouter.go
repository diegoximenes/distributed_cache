package httprouter

import (
	"fmt"
	"net/http"

	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/config"
	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/httprouter/handlers/node"
	raftHandler "github.com/diegoximenes/distributed_cache/nodemetadata/internal/httprouter/handlers/raft"
	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/fsm"
	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/metadata"
	raftMetadata "github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/metadata"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
)

func checkRaftLeaderMiddleware(raftNode *raft.Raft, raftNodeMetadataClient *metadata.RaftNodeMetadataClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := raftNode.VerifyLeader().Error()
		if err != nil {
			leaderApplicationAddress, err := raftNodeMetadataClient.GetLeaderApplicationAddress()
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
			} else {
				url := fmt.Sprintf("%v%v", leaderApplicationAddress, c.Request.URL.Path)
				c.Redirect(http.StatusTemporaryRedirect, url)
			}
			c.Abort()
		}
	}
}

func Set(raftNode *raft.Raft, fsm *fsm.FSM, raftNodeMetadataClient *raftMetadata.RaftNodeMetadataClient) {
	router := gin.Default()

	raftLeaderGroup := router.Group("/", checkRaftLeaderMiddleware(raftNode, raftNodeMetadataClient))
	raftLeaderGroup.GET("/node", node.Get(raftNode, fsm, raftNodeMetadataClient))
	raftLeaderGroup.PUT("/node", node.Put(raftNode, raftNodeMetadataClient))
	raftLeaderGroup.DELETE("/node/:id", node.Delete(raftNode, raftNodeMetadataClient))
	raftLeaderGroup.PUT("/raft/join", raftHandler.Join(raftNode))
	raftLeaderGroup.GET("/raft/nodes", raftHandler.Nodes(raftNodeMetadataClient))

	router.Run(config.Config.ApplicationAddress)
}
