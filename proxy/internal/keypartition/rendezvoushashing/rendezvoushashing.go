package rendezvoushashing

import (
	"fmt"

	keyPartitionErrors "github.com/diegoximenes/distributed_cache/proxy/internal/keypartition/errors"
	"github.com/spaolacci/murmur3"
)

type RendezvousHashing struct {
	nodesID []string
}

func (rendezvousHashing *RendezvousHashing) UpdateNodes(nodesID []string) {
	copiedNodesID := make([]string, len(nodesID))
	copy(copiedNodesID, nodesID)

	rendezvousHashing.nodesID = copiedNodesID
}

func (rendezvousHashing *RendezvousHashing) GetNodeID(
	objKey string,
) (string, error) {
	if len(rendezvousHashing.nodesID) == 0 {
		return "", keyPartitionErrors.NoAvailableNodesError
	}

	bestHash := uint64(0)
	bestNodeID := ""
	for _, nodeID := range rendezvousHashing.nodesID {
		hash := murmur3.Sum64([]byte(fmt.Sprintf("%s:%s", nodeID, objKey)))
		if bestHash < hash {
			bestNodeID = nodeID
			bestHash = hash
		}
	}

	return bestNodeID, nil
}
