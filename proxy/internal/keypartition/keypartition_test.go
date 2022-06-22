package keypartition

import (
	"math/rand"
	"testing"
	"time"

	"github.com/diegoximenes/distributed_cache/proxy/internal/keypartition/rendezvoushashing"
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

func testKeyPartitionStrategy(keyPartitionStrategy KeyPartitionStrategy, nodesID []string) {
	keyPartitionStrategy.UpdateNodes(nodesID)

	Convey("uniform distribution", func() {
		numberOfKeys := 100000
		countSelectedNodeID := make(map[string]int)
		for i := 0; i < numberOfKeys; i++ {
			objKeyLen := rand.Intn(100) + 1
			objKey := randString(objKeyLen)

			selectedNodeID, err := keyPartitionStrategy.GetNodeID(objKey)
			So(err, ShouldBeNil)

			countSelectedNodeID[selectedNodeID] += 1
		}

		for _, nodeID := range nodesID {
			fractionOfKeysInNode :=
				float64(countSelectedNodeID[nodeID]) / float64(numberOfKeys)
			So(fractionOfKeysInNode, ShouldBeBetween, 0.24, 0.26)
		}
	})

	Convey("after a node removal only keys that were forwarded to this node are moved to other nodes", func() {
		numberOfKeys := 100000
		objsKey := make([]string, numberOfKeys)
		for i := 0; i < numberOfKeys; i++ {
			objKeyLen := rand.Intn(100) + 1
			objsKey[i] = randString(objKeyLen)
		}

		objKeyToNodeIDBeforeRemoval := make(map[string]string)
		for _, objKey := range objsKey {
			selectedNodeID, err := keyPartitionStrategy.GetNodeID(objKey)
			So(err, ShouldBeNil)
			objKeyToNodeIDBeforeRemoval[objKey] = selectedNodeID
		}

		nodeIDToBeRemoved := "node2"
		nodesIDAfterRemoval := make([]string, len(nodesID)-1)
		i := 0
		for _, nodeID := range nodesID {
			if nodeID != nodeIDToBeRemoved {
				nodesIDAfterRemoval[i] = nodeID
				i++
			}
		}
		keyPartitionStrategy.UpdateNodes(nodesIDAfterRemoval)

		for _, objKey := range objsKey {
			selectedNodeID, err := keyPartitionStrategy.GetNodeID(objKey)
			So(err, ShouldBeNil)
			if objKeyToNodeIDBeforeRemoval[objKey] != nodeIDToBeRemoved {
				So(objKeyToNodeIDBeforeRemoval[objKey], ShouldEqual, selectedNodeID)
			} else {
				So(objKeyToNodeIDBeforeRemoval[objKey], ShouldNotEqual, selectedNodeID)
			}
		}
	})
}

func TestKeyPartition(test *testing.T) {
	Convey("Setup", test, func() {
		rand.Seed(time.Now().UnixNano())

		nodesID := []string{"node0", "node1", "node2", "node3"}

		Convey("RendezvousHashing", func() {
			rendezvouzHashing := &rendezvoushashing.RendezvousHashing{}
			testKeyPartitionStrategy(rendezvouzHashing, nodesID)
		})
	})
}
