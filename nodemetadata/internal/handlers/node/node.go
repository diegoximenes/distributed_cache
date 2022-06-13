package node

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"

	raftFSM "github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/raft/fsm"
)

const (
	raftTimeout = 10 * time.Second
)

func Get(raftNode *raft.Raft, fsm *raftFSM.FSM) func(c *gin.Context) {
	return func(c *gin.Context) {
		err := raftNode.VerifyLeader().Error()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		} else {
			nodesMetadata := fsm.Get()
			c.JSON(http.StatusOK, nodesMetadata)
		}
	}
}

func applyCommand(raftNode *raft.Raft, command *raftFSM.Command) error {
	commandBytes, err := json.Marshal(command)
	if err != nil {
		return err
	}
	return raftNode.Apply(commandBytes, raftTimeout).Error()
}

func Put(raftNode *raft.Raft) func(c *gin.Context) {
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

func Delete(raftNode *raft.Raft) func(c *gin.Context) {
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