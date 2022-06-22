package keypartition

import "github.com/diegoximenes/distributed_cache/proxy/internal/keypartition/rendezvoushashing"

type KeyPartitionStrategy interface {
	UpdateNodes(nodesID []string)
	GetNodeID(objKey string) (string, error)
}

func New() KeyPartitionStrategy {
	return &rendezvoushashing.RendezvousHashing{}
}
