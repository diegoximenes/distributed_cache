package httprouter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/config"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/httprouter/handlers/nodes"
	raftHandler "github.com/diegoximenes/distributed_cache/nodesmetadata/internal/httprouter/handlers/raft"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/logger"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/fsm"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/metadata"
	raftMetadata "github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/metadata"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/pkg/net/sse"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"go.uber.org/zap"
)

func getIPv4AndPort(ctx context.Context, hostAndPort string) (string, string, error) {
	host, port, err := net.SplitHostPort(hostAndPort)
	if err != nil {
		return "", "", err
	}

	ip, err := net.DefaultResolver.LookupIP(ctx, "ip4", host)
	if err != nil {
		return "", "", err
	}

	return ip[0].String(), port, nil
}

func checkRaftLeaderMiddleware(
	raftNode *raft.Raft,
	raftMetadataClient *metadata.RaftMetadataClient,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := raftNode.VerifyLeader().Error()
		if err != nil {
			defer c.Abort()

			leaderApplicationAddress, err :=
				raftMetadataClient.GetLeaderApplicationAddress(c.Request.Context())
			if err != nil {
				if errors.Is(err, context.Canceled) {
					// maybe useful for metrics tracking purposes in the server,
					// but is useful to not log as 5xx
					c.AbortWithStatus(499)
					return
				}

				logger.Logger.Error(
					err.Error(),
					zap.String("middleware", "checkRaftLeaderMiddleware"),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			// use IP to handle the following case: nodesmetadata running as docker
			// containers using the same bridge network, being able to communicate
			// with each other through names. The host of these containers is
			// not able to resolve those names.
			leaderIP, leaderApplicationPort, err :=
				getIPv4AndPort(c.Request.Context(), leaderApplicationAddress)
			if err != nil {
				logger.Logger.Error(
					err.Error(),
					zap.String("middleware", "checkRaftLeaderMiddleware"),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			url := fmt.Sprintf(
				"http://%v:%v%v",
				leaderIP,
				leaderApplicationPort,
				c.Request.URL.Path,
			)
			c.Redirect(http.StatusTemporaryRedirect, url)
		}
	}
}

func Set(
	raftNode *raft.Raft,
	fsm *fsm.FSM,
	raftMetadataClient *raftMetadata.RaftMetadataClient,
) {
	nodesSSE := sse.New()
	raftMetadataSSE := metadata.NewSSE(raftNode, nodesSSE)

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

	router.Run(config.Config.ApplicationBindAddress)
}
