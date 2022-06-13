package rendezvoushashing

import (
	"fmt"

	"github.com/spaolacci/murmur3"

	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/nodemetadata"
)

func GetNodeMetadata(nodesMetadata *nodemetadata.NodesMetadata, key string) *nodemetadata.NodeMetadata {
	bestH1 := uint64(0)
	bestH2 := uint64(0)
	bestNodeMetadata := &nodemetadata.NodeMetadata{}

	if len(*nodesMetadata) == 0 {
		return nil
	}

	for nodeId, nodeMetadata := range *nodesMetadata {
		h1, h2 := murmur3.Sum128([]byte(fmt.Sprintf("%v:%v", nodeId, key)))
		if (bestH2 < h2) || ((bestH2 == h2) && (bestH1 <= h1)) {
			bestNodeMetadata = &nodeMetadata
			bestH1 = h1
			bestH2 = h2
		}
	}

	return bestNodeMetadata
}
