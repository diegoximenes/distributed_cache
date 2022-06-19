package membership

import (
	"time"

	"github.com/hashicorp/raft"
)

type AddInput struct {
	ID      string `json:"id" binding:"required"`
	Address string `json:"address" binding:"required"`
}

func Add(raftNode *raft.Raft, input *AddInput) error {
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

func Remove(raftNode *raft.Raft, nodeID string) error {
	return raftNode.
		RemoveServer(raft.ServerID(nodeID), 0, 2*time.Second).
		Error()
}
