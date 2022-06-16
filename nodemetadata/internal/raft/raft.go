package raft

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"

	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/config"
	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/fsm"
	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/metadata"
	"github.com/diegoximenes/distributed_cache/nodemetadata/internal/raft/streamlayer"
	"github.com/diegoximenes/distributed_cache/nodemetadata/pkg/net/connection/demux"
	"github.com/diegoximenes/distributed_cache/nodemetadata/pkg/net/connection/listener"
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
	transport := raft.NewNetworkTransport(streamLayer, 5, 10*time.Second, nil)
	return transport, nil
}

func setRaftNodeMetadataServer(demux *demux.Demux, tcpAddr *net.TCPAddr) {
	raftNodeMetadataListener := listener.New(tcpAddr)
	demux.RegisterOutListener(raftNodeMetadataFirstByte, raftNodeMetadataListener)

	router := gin.Default()
	router.GET(metadata.HTTPPath, func(c *gin.Context) {
		response := metadata.Response{
			ApplicationAddress: config.Config.ApplicationAddress,
		}
		c.JSON(http.StatusOK, response)
	})
	go http.Serve(raftNodeMetadataListener, router)
	// TODO: error handling
}

func Set() (*raft.Raft, *fsm.FSM, *metadata.RaftNodeMetadataClient, error) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(config.Config.RaftId)

	fsm := fsm.New()

	logsStore, err := raftboltdb.NewBoltStore(filepath.Join(config.Config.RaftDir, "logs.dat"))
	if err != nil {
		return nil, nil, nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(config.Config.RaftDir, "stable.dat"))
	if err != nil {
		return nil, nil, nil, err
	}

	snapshotsStore, err := raft.NewFileSnapshotStore(config.Config.RaftDir, 2, os.Stderr)
	if err != nil {
		return nil, nil, nil, err
	}

	demux, err := demux.New(config.Config.RaftAddress)
	if err != nil {
		return nil, nil, nil, err
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", config.Config.RaftAddress)
	if err != nil {
		return nil, nil, nil, err
	}
	transport, err := getTransport(demux, tcpAddr)
	if err != nil {
		return nil, nil, nil, err
	}

	raftNode, err := raft.NewRaft(raftConfig, fsm, logsStore, stableStore, snapshotsStore, transport)
	if err != nil {
		return nil, nil, nil, err
	}

	setRaftNodeMetadataServer(demux, tcpAddr)
	raftNodeMetadataClient := metadata.NewClient(raftNode, raftNodeMetadataFirstByte)

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

	return raftNode, fsm, raftNodeMetadataClient, nil
}
