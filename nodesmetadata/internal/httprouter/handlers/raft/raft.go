package raft

import (
	"net/http"

	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/httprouter/handlers"
	raftMembership "github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/membership"
	raftMetadata "github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/metadata"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
)

func PutNode(raftNode *raft.Raft) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input raftMembership.AddInput
		err := c.BindJSON(&input)
		if err != nil {
			c.JSON(http.StatusBadRequest, handlers.APIError{Error: err.Error()})
			return
		}

		err = raftMembership.Add(raftNode, &input)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.AbortWithStatus(http.StatusOK)
	}
}

func DeleteNode(raftNode *raft.Raft) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		err := raftMembership.Remove(raftNode, id)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.AbortWithStatus(http.StatusOK)
	}
}

func Metadata(raftMetadataClient *raftMetadata.RaftMetadataClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		raftMetadata :=
			raftMetadataClient.GetRaftMetadata(c.Request.Context())
		c.JSON(http.StatusOK, raftMetadata)
	}
}
