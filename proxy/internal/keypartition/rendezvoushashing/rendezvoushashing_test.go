package rendezvoushashing

import (
	"math/rand"
	"testing"
	"time"

	"github.com/diegoximenes/distributed_cache/proxy/pkg/clients/nodesmetadata"
	. "github.com/smartystreets/goconvey/convey"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestRendezvousHashing(test *testing.T) {
	Convey("Setup", test, func() {
		rand.Seed(time.Now().UnixNano())
		nodesMetadata := make(nodesmetadata.NodesMetadata)
		nodesMetadata["node0"] = nodesmetadata.NodeMetadata{
			ID:      "node0",
			Address: "http://localhost:30000",
		}
		nodesMetadata["node1"] = nodesmetadata.NodeMetadata{
			ID:      "node1",
			Address: "http://localhost:30001",
		}
		nodesMetadata["node2"] = nodesmetadata.NodeMetadata{
			ID:      "node2",
			Address: "http://localhost:30002",
		}
		nodesMetadata["node3"] = nodesmetadata.NodeMetadata{
			ID:      "node3",
			Address: "http://localhost:30003",
		}

		Convey("uniform distribution", func() {
			numberOfKeys := 100000
			countSelectedNodeMetadata := make(map[string]int)
			for i := 0; i < numberOfKeys; i++ {
				keyLen := rand.Intn(100) + 1
				key := randString(keyLen)

				selectedNodeMetadata := GetNodeMetadata(&nodesMetadata, key)
				countSelectedNodeMetadata[selectedNodeMetadata.ID] += 1
			}

			for nodeId := range nodesMetadata {
				fractionOfKeysInNode :=
					float64(countSelectedNodeMetadata[nodeId]) / float64(numberOfKeys)
				So(fractionOfKeysInNode, ShouldBeBetween, 0.24, 0.26)
			}
		})

		Convey("after a node removal only keys that were forwarded to this node are moved to other nodes", func() {
			numberOfKeys := 100000
			keys := make([]string, numberOfKeys)
			for i := 0; i < numberOfKeys; i++ {
				keyLen := rand.Intn(100) + 1
				keys[i] = randString(keyLen)
			}

			keyToNodeIDBeforeRemoval := make(map[string]string)
			for _, key := range keys {
				selectedNodeMetadata := GetNodeMetadata(&nodesMetadata, key)
				keyToNodeIDBeforeRemoval[key] = selectedNodeMetadata.ID
			}

			nodeIDToBeRemoved := "node2"
			delete(nodesMetadata, nodeIDToBeRemoved)

			for _, key := range keys {
				selectedNodeMetadata := GetNodeMetadata(&nodesMetadata, key)
				if keyToNodeIDBeforeRemoval[key] != nodeIDToBeRemoved {
					So(keyToNodeIDBeforeRemoval[key], ShouldEqual, selectedNodeMetadata.ID)
				} else {
					So(keyToNodeIDBeforeRemoval[key], ShouldNotEqual, selectedNodeMetadata.ID)
				}
			}
		})
	})
}
