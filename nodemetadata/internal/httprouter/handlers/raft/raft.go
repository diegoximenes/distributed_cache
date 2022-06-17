package raft

import (
	"net/http"

	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/httprouter/handlers"
	raftJoin "github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/join"
	raftMetadata "github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/metadata"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
)

func Join(raftNode *raft.Raft) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input raftJoin.JoinInput
		err := c.BindJSON(&input)
		if err != nil {
			c.JSON(http.StatusBadRequest, handlers.APIError{Error: err.Error()})
			return
		}

		err = raftJoin.Join(raftNode, &input)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.AbortWithStatus(http.StatusOK)
	}
}

type NodesResponse struct {
	NodesApplicationAddresses []string `json:"nodesApplicationAddresses"`
}

func Nodes(raftNodeMetadataClient *raftMetadata.RaftNodeMetadataClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		response := NodesResponse{
			NodesApplicationAddresses: raftNodeMetadataClient.GetNodesApplicationAddresses(),
		}
		c.JSON(http.StatusOK, response)
	}
}
