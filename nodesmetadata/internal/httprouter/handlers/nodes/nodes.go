package nodes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"

	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/httprouter/handlers"
	raftFSM "github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/fsm"
)

const (
	raftTimeout = 2 * time.Second
)

func Get(raftNode *raft.Raft, fsm *raftFSM.FSM) func(c *gin.Context) {
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

func Put(raftNode *raft.Raft) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input raftFSM.NodeMetadata
		err := c.BindJSON(&input)
		if err != nil {
			apiError := handlers.APIError{
				Error: err.Error(),
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, apiError)
			return
		}

		command := raftFSM.Command{
			Operation:    "set",
			NodeMetadata: input,
		}
		err = applyCommand(raftNode, &command)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.AbortWithStatus(http.StatusOK)
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
