package rendezvoushashing

import (
	"fmt"

	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/nodesmetadata"
	"github.com/spaolacci/murmur3"
)

func GetNodeMetadata(
	nodesMetadata *nodesmetadata.NodesMetadata,
	key string,
) *nodesmetadata.NodeMetadata {
	if len(*nodesMetadata) == 0 {
		return nil
	}

	bestHash := uint64(0)
	bestNodeId := ""
	for nodeId := range *nodesMetadata {
		hash := murmur3.Sum64([]byte(fmt.Sprintf("%s:%s", nodeId, key)))
		if bestHash < hash {
			bestNodeId = nodeId
			bestHash = hash
		}
	}

	bestNodeMetadata := (*nodesMetadata)[bestNodeId]
	return &bestNodeMetadata
}
