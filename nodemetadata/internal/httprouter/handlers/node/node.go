package node

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"

	raftFSM "github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/fsm"
	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/metadata"
)

const (
	raftTimeout = 10 * time.Second
)

func Get(raftNode *raft.Raft, fsm *raftFSM.FSM, raftNodeMetadataClient *metadata.RaftNodeMetadataClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		nodesMetadata := fsm.Get()
		c.JSON(http.StatusOK, nodesMetadata)
	}
}

func applyCommand(raftNode *raft.Raft, command *raftFSM.Command) error {
	commandBytes, err := json.Marshal(command)
	if err != nil {
		return err
	}
	return raftNode.Apply(commandBytes, raftTimeout).Error()
}

func Put(raftNode *raft.Raft, raftNodeMetadataClient *metadata.RaftNodeMetadataClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input raftFSM.NodeMetadata
		c.BindJSON(&input)

		command := raftFSM.Command{
			Operation:    "set",
			NodeMetadata: input,
		}
		err := applyCommand(raftNode, &command)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		} else {
			c.AbortWithStatus(http.StatusOK)
		}
	}
}

func Delete(raftNode *raft.Raft, raftNodeMetadataClient *metadata.RaftNodeMetadataClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		command := raftFSM.Command{
			Operation: "delete",
			NodeMetadata: raftFSM.NodeMetadata{
				ID: id,
			},
		}
		err := applyCommand(raftNode, &command)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		} else {
			c.AbortWithStatus(http.StatusOK)
		}
	}
}
