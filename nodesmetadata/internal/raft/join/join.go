package join

import (
	"time"

	"github.com/hashicorp/raft"
)

type JoinInput struct {
	ID      string `json:"id" binding:"required"`
	Address string `json:"address" binding:"required"`
}

func Join(raftNode *raft.Raft, input *JoinInput) error {
	err := raftNode.
		AddVoter(
			raft.ServerID(input.ID),
			raft.ServerAddress(input.Address),
			0,
			2*time.Second,
		).
		Error()
	return err
}
