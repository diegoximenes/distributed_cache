package consistenthashing

import (
	"fmt"

	keyPartitionErrors "github.com/diegoximenes/distributed_cache/proxy/internal/keypartition/errors"
	"github.com/emirpasic/gods/trees/redblacktree"
	godsUtils "github.com/emirpasic/gods/utils"
	"github.com/spaolacci/murmur3"
)

type ConsistentHashing struct {
	nodesID *redblacktree.Tree
}

const (
	numberOfVirtualNodesPerRealNode = 200
)

func (consistentHashing *ConsistentHashing) hash(s string) uint64 {
	return murmur3.Sum64([]byte(s))
}

func (consistentHashing *ConsistentHashing) UpdateNodes(nodesID []string) {
	nodesIDTree := redblacktree.NewWith(godsUtils.UInt64Comparator)
	for _, nodeID := range nodesID {
		for virtualNode := 0; virtualNode < numberOfVirtualNodesPerRealNode; virtualNode++ {
			nodeKey := fmt.Sprintf("%v:%v", nodeID, virtualNode)
			nodeKeyHash := consistentHashing.hash(nodeKey)
			// don't handle collisions
			nodesIDTree.Put(nodeKeyHash, nodeID)
		}
	}

	consistentHashing.nodesID = nodesIDTree
}

func (consistentHashing *ConsistentHashing) GetNodeID(
	objKey string,
) (string, error) {
	// don't operates directly on consistentHashing.nodesID since it can change
	// in the middle of this function execution
	nodesID := consistentHashing.nodesID

	if nodesID.Empty() {
		return "", keyPartitionErrors.NoAvailableNodesError
	}

	objKeyHash := consistentHashing.hash(objKey)

	node, found := nodesID.Ceiling(objKeyHash)
	if !found {
		node = nodesID.Left()
	}

	return fmt.Sprintf("%v", node.Value), nil
}
