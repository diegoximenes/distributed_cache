package rendezvoushashing

import (
	"fmt"

	"github.com/spaolacci/murmur3"

	"github.com/diegoximenes/distributed_key_value_cache/proxy/pkg/clients/configserver"
)

func GetNodeConfig(nodesConfig *configserver.NodesConfig, key string) *configserver.NodeConfig {
	bestH1 := uint64(0)
	bestH2 := uint64(0)
	bestNodeConfig := &configserver.NodeConfig{}

	for nodeId, nodeConfig := range *nodesConfig {
		h1, h2 := murmur3.Sum128([]byte(fmt.Sprintf("%v:%v", nodeId, key)))
		if (bestH2 < h2) || ((bestH2 == h2) && (bestH1 <= h1)) {
			bestNodeConfig = &nodeConfig
			bestH1 = h1
			bestH2 = h2
		}
	}

	return bestNodeConfig
}
