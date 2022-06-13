package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type NodeMetadata struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

type NodesMetadata map[string]NodeMetadata

type Command struct {
	Operation    string       `json:"operation,omitempty"`
	NodeMetadata NodeMetadata `json:"nodeMetadata,omitempty"`
}

type FSM struct {
	rwMutex       sync.RWMutex
	nodesMetadata NodesMetadata
}

func New() *FSM {
	return &FSM{
		nodesMetadata: make(NodesMetadata),
	}
}

func (fsm *FSM) applySet(nodeMetadata *NodeMetadata) interface{} {
	fsm.rwMutex.Lock()
	defer fsm.rwMutex.Unlock()
	fsm.nodesMetadata[nodeMetadata.ID] = *nodeMetadata
	return nil
}

func (fsm *FSM) applyDelete(nodeMetadata *NodeMetadata) interface{} {
	fsm.rwMutex.Lock()
	defer fsm.rwMutex.Unlock()
	delete(fsm.nodesMetadata, nodeMetadata.ID)
	return nil
}

func cloneNodesMetadata(toBeCloned NodesMetadata) NodesMetadata {
	cloned := make(NodesMetadata)
	for k, v := range toBeCloned {
		cloned[k] = v
	}
	return cloned
}

func (fsm *FSM) Apply(log *raft.Log) interface{} {
	var command Command
	if err := json.Unmarshal(log.Data, &command); err != nil {
		panic(err)
	}

	switch command.Operation {
	case "set":
		return fsm.applySet(&command.NodeMetadata)
	case "delete":
		return fsm.applyDelete(&command.NodeMetadata)
	default:
		panic(fmt.Sprintf("Unrecognized command operation: %v", command.Operation))
	}
}

func (fsm *FSM) Snapshot() (raft.FSMSnapshot, error) {
	fsm.rwMutex.Lock()
	defer fsm.rwMutex.Unlock()

	clonedNodesMetadata := cloneNodesMetadata(fsm.nodesMetadata)

	return &fsmSnapshot{nodesMetadata: clonedNodesMetadata}, nil
}

func (fsm *FSM) Restore(rc io.ReadCloser) error {
	nodesMetadata := make(NodesMetadata)
	if err := json.NewDecoder(rc).Decode(&nodesMetadata); err != nil {
		return err
	}
	fsm.nodesMetadata = nodesMetadata
	return nil
}

func (fsm *FSM) Get() NodesMetadata {
	fsm.rwMutex.RLock()
	defer fsm.rwMutex.RUnlock()

	return cloneNodesMetadata(fsm.nodesMetadata)
}
