package raft

import (
	"net/http"

	raftJoin "github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/join"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
)

func Join(raftNode *raft.Raft) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input raftJoin.JoinInput
		c.BindJSON(&input)

		err := raftJoin.Join(raftNode, &input)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		} else {
			c.AbortWithStatus(http.StatusOK)
		}
	}
}
