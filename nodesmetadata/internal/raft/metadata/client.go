package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/diegoximenes/distributed_cache/nodesmetadata/pkg/net/connection/mux"
	httpUtil "github.com/diegoximenes/distributed_cache/util/pkg/http"
	"github.com/hashicorp/raft"
	"golang.org/x/net/context"
)

type RaftNodeMetadata struct {
	NodeID             string `json:"nodeID"`
	ApplicationAddress string `json:"applicationAddress"`
	RaftAddress        string `json:"raftAddress"`
}

type RaftNodesMetadata map[string]*RaftNodeMetadata

type RaftMetadata struct {
	NodesMetadata RaftNodesMetadata `json:"nodesMetadata"`
	LeaderNodeID  string            `json:"leaderNodeID"`
}

type RaftMetadataClient struct {
	httpClient *httpUtil.HTTPClient
	raftNode   *raft.Raft
}

func NewClient(raftNode *raft.Raft, firstByte byte) *RaftMetadataClient {
	transport := &http.Transport{
		Dial: func(network string, address string) (net.Conn, error) {
			return mux.Dial(network, address, 1*time.Second, firstByte)
		},
	}
	httpClient := httpUtil.NewClient(&http.Client{
		Transport: transport,
		Timeout:   1 * time.Second,
	})

	return &RaftMetadataClient{
		httpClient: httpClient,
		raftNode:   raftNode,
	}
}

func (client *RaftMetadataClient) getApplicationAddress(
	ctx context.Context,
	raftAddress string,
) (string, error) {
	url := fmt.Sprintf("http://%v%v", raftAddress, HTTPPath)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	responseBytes, err := client.httpClient.DoRequest(request)
	if err != nil {
		return "", err
	}
	var response Response
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return "", nil
	}

	applicationAddress := response.ApplicationAddress
	return applicationAddress, nil
}

func (client *RaftMetadataClient) GetLeaderApplicationAddress(
	ctx context.Context,
) (string, error) {
	leaderRaftAddress, _ := client.raftNode.LeaderWithID()
	if leaderRaftAddress == "" {
		return "", errors.New("unknown leader")
	}
	return client.getApplicationAddress(ctx, string(leaderRaftAddress))
}

func (client *RaftMetadataClient) addRaftNodeMetadataToChannel(
	ctx context.Context,
	raftServer *raft.Server,
	raftNodeMetadataChannel chan *RaftNodeMetadata,
) {
	raftNodeMetadata := RaftNodeMetadata{
		NodeID:             string(raftServer.ID),
		RaftAddress:        string(raftServer.Address),
		ApplicationAddress: "",
	}

	applicationAddress, err := client.getApplicationAddress(ctx, string(raftServer.Address))
	if err == nil {
		raftNodeMetadata.ApplicationAddress = applicationAddress
	}

	raftNodeMetadataChannel <- &raftNodeMetadata
}

func (client *RaftMetadataClient) GetRaftMetadata(
	ctx context.Context,
) *RaftMetadata {
	nodes := client.raftNode.GetConfiguration().Configuration().Servers

	raftNodeMetadataChannel := make(chan *RaftNodeMetadata, len(nodes))
	for i := range nodes {
		go client.addRaftNodeMetadataToChannel(
			ctx,
			&nodes[i],
			raftNodeMetadataChannel,
		)
	}

	raftNodesMetadata := make(RaftNodesMetadata)
	for i := 0; i < len(nodes); i++ {
		raftNodeMetadata := <-raftNodeMetadataChannel
		raftNodesMetadata[raftNodeMetadata.NodeID] = raftNodeMetadata
	}

	_, leaderID := client.raftNode.LeaderWithID()

	raftMetadata := RaftMetadata{
		NodesMetadata: raftNodesMetadata,
		LeaderNodeID:  string(leaderID),
	}

	return &raftMetadata
}
