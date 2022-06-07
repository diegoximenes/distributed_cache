package rendezvoushashing

import (
	"fmt"
	"sort"

	"github.com/spaolacci/murmur3"

	"github.com/diegoximenes/distributed_key_value_cache/proxy/pkg/clients/configserver"
)

type nodeScore struct {
	id string
	h1 uint64
	h2 uint64
}

type nodesScore []nodeScore

func (n nodesScore) Len() int {
	return len(n)
}

func (n nodesScore) Less(i, j int) bool {
	if n[i].h2 < n[j].h2 {
		return true
	} else if n[i].h2 > n[j].h2 {
		return false
	}
	return n[i].h1 < n[j].h1
}

func (n nodesScore) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func GetNodeConfig(nodesConfig *configserver.NodesConfig, key string) *configserver.NodeConfig {
	scores := nodesScore{}
	for nodeId := range *nodesConfig {
		h1, h2 := murmur3.Sum128([]byte(fmt.Sprintf("%v:%v", nodeId, key)))
		scores = append(scores, nodeScore{
			id: nodeId,
			h1: h1,
			h2: h2,
		})
	}
	sort.Sort(nodesScore(scores))

	nodeConfig := (*nodesConfig)[scores[len(scores)-1].id]
	return &nodeConfig
}
