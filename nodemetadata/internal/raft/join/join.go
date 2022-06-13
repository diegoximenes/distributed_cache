package join

import (
	"time"

	"github.com/hashicorp/raft"
)

type JoinInput struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

func Join(raftNode *raft.Raft, input *JoinInput) error {
	return raftNode.AddVoter(raft.ServerID(input.ID), raft.ServerAddress(input.Address), 0, 5*time.Second).Error()
}
