package keypartition

import (
	"errors"

	"github.com/diegoximenes/distributed_cache/proxy/internal/config"
	"github.com/diegoximenes/distributed_cache/proxy/internal/keypartition/consistenthashing"
	"github.com/diegoximenes/distributed_cache/proxy/internal/keypartition/rendezvoushashing"
)

var NoNodesAvailableError = errors.New("no nodes are available")

type KeyPartitionStrategy interface {
	UpdateNodes(nodesID []string)
	GetNodeID(objKey string) (string, error)
}

func New() KeyPartitionStrategy {
	switch config.Config.KeyPartitionStrategy {
	case config.ConsistentHashing:
		return &consistenthashing.ConsistentHashing{}
	default:
		return &rendezvoushashing.RendezvousHashing{}
	}
}
