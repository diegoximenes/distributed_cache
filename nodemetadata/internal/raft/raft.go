package raft

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"

	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/config"
	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/raft/fsm"
)

func Set() (*raft.Raft, *fsm.FSM, error) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(config.Config.RaftId)

	fsm := fsm.New()

	logsStore, err := raftboltdb.NewBoltStore(filepath.Join(config.Config.RaftDir, "logs.dat"))
	if err != nil {
		return nil, nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(config.Config.RaftDir, "stable.dat"))
	if err != nil {
		return nil, nil, err
	}

	snapshotsStore, err := raft.NewFileSnapshotStore(config.Config.RaftDir, 2, os.Stderr)
	if err != nil {
		return nil, nil, err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", config.Config.RaftAddress)
	if err != nil {
		return nil, nil, err
	}
	transport, err := raft.NewTCPTransport(config.Config.RaftAddress, tcpAddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		fmt.Printf(err.Error())
		return nil, nil, err
	}

	raftNode, err := raft.NewRaft(raftConfig, fsm, logsStore, stableStore, snapshotsStore, transport)
	if err != nil {
		return nil, nil, err
	}

	if *config.Config.BootstrapRaftCluster {
		clusterConfig := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		raftNode.BootstrapCluster(clusterConfig)
	}

	return raftNode, fsm, nil
}
