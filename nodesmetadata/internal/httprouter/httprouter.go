package httprouter

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/config"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/httprouter/handlers/nodes"
	raftHandler "github.com/diegoximenes/distributed_cache/nodesmetadata/internal/httprouter/handlers/raft"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/fsm"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/metadata"
	raftMetadata "github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/metadata"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/pkg/net/sse"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
)

func checkRaftLeaderMiddleware(
	raftNode *raft.Raft,
	raftMetadataClient *metadata.RaftMetadataClient,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := raftNode.VerifyLeader().Error()
		if err != nil {
			leaderApplicationAddress, err :=
				raftMetadataClient.GetLeaderApplicationAddress(c.Request.Context())
			if err != nil {
				if errors.Is(err, context.Canceled) {
					// maybe useful for metrics tracking purposes in the server,
					// but is useful to not log as 5xx
					c.AbortWithStatus(499)
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			} else {
				url := fmt.Sprintf("%v%v", leaderApplicationAddress, c.Request.URL.Path)
				c.Redirect(http.StatusTemporaryRedirect, url)
			}
			c.Abort()
		}
	}
}

func Set(
	raftNode *raft.Raft,
	fsm *fsm.FSM,
	raftMetadataClient *raftMetadata.RaftMetadataClient,
) {
	nodesSSE := sse.New()
	raftMetadataSSE := metadata.NewSSE(raftNode)

	router := gin.Default()

	raftLeaderGroup :=
		router.Group("/", checkRaftLeaderMiddleware(raftNode, raftMetadataClient))

	raftLeaderGroup.GET("/nodes", nodes.Get(raftNode, fsm))
	raftLeaderGroup.PUT("/nodes", nodes.Put(nodesSSE.EventsToSend, raftNode))
	raftLeaderGroup.DELETE("/nodes/:id", nodes.Delete(nodesSSE.EventsToSend, raftNode))
	raftLeaderGroup.GET("/nodes/sse", nodesSSE.Handler())

	raftLeaderGroup.PUT("/raft/node", raftHandler.PutNode(raftNode))
	raftLeaderGroup.DELETE("/raft/node/:id", raftHandler.DeleteNode(raftNode))
	raftLeaderGroup.GET("/raft/metadata", raftHandler.Metadata(raftMetadataClient))
	raftLeaderGroup.GET("/raft/metadata/sse", raftMetadataSSE.Handler())

	router.Run(config.Config.ApplicationAddress)
}
