package raft

import (
	"net"
	"os"
	"path/filepath"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"

	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/config"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/fsm"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/metadata"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/streamlayer"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/raft/timeout"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/pkg/net/connection/demux"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/pkg/net/connection/listener"
)

const (
	raftProtocolFirstByte     = byte(1)
	raftNodeMetadataFirstByte = byte(2)
)

func getTransport(demux *demux.Demux, tcpAddr *net.TCPAddr) (*raft.NetworkTransport, error) {
	raftProtocolListener := listener.New(tcpAddr)
	err := demux.RegisterOutListener(raftProtocolFirstByte, raftProtocolListener)
	if err != nil {
		return nil, err
	}
	streamLayer := streamlayer.New(raftProtocolListener, raftProtocolFirstByte)
	transport := raft.NewNetworkTransport(streamLayer, 5, timeout.RaftTimeout, nil)
	return transport, nil
}

func Set() (*raft.Raft, *fsm.FSM, *metadata.RaftMetadataClient, error) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(config.Config.RaftID)

	fsm := fsm.New()

	logsStore, err := raftboltdb.NewBoltStore(filepath.Join(config.Config.RaftDir, "logs.dat"))
	if err != nil {
		return nil, nil, nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(config.Config.RaftDir, "stable.dat"))
	if err != nil {
		return nil, nil, nil, err
	}

	snapshotsStore, err := raft.NewFileSnapshotStore(config.Config.RaftDir, 2, os.Stdout)
	if err != nil {
		return nil, nil, nil, err
	}

	demux, err := demux.New(config.Config.RaftBindAddress)
	if err != nil {
		return nil, nil, nil, err
	}
	tcpRaftAdvertisedAddr, err := net.ResolveTCPAddr("tcp", config.Config.RaftAdvertisedAddress)
	if err != nil {
		return nil, nil, nil, err
	}
	transport, err := getTransport(demux, tcpRaftAdvertisedAddr)
	if err != nil {
		return nil, nil, nil, err
	}

	raftNode, err := raft.NewRaft(raftConfig, fsm, logsStore, stableStore, snapshotsStore, transport)
	if err != nil {
		return nil, nil, nil, err
	}

	metadata.SetServer(demux, tcpRaftAdvertisedAddr, raftNodeMetadataFirstByte)
	raftMetadataClient := metadata.NewClient(raftNode, raftNodeMetadataFirstByte)

	if config.Config.BootstrapRaftCluster {
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

	return raftNode, fsm, raftMetadataClient, nil
}
