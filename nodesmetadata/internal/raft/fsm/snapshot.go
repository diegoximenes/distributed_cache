package fsm

import (
	"encoding/json"

	"github.com/hashicorp/raft"
)

type fsmSnapshot struct {
	nodesMetadata NodesMetadata
}

func (fs *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	bytes, err := json.Marshal(fs.nodesMetadata)
	if err != nil {
		sink.Cancel()
		return err
	}

	if _, err := sink.Write(bytes); err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (fs *fsmSnapshot) Release() {}
