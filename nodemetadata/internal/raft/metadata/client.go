package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/diegoximenes/distributed_cache/nodemetadata/pkg/net/connection/mux"
	httpUtil "github.com/diegoximenes/distributed_cache/util/pkg/http"
	"github.com/hashicorp/raft"
)

type RaftNodeMetadataClient struct {
	raftNode  *raft.Raft
	firstByte byte
}

func NewClient(raftNode *raft.Raft, firstByte byte) *RaftNodeMetadataClient {
	return &RaftNodeMetadataClient{
		raftNode:  raftNode,
		firstByte: firstByte,
	}
}

func (client *RaftNodeMetadataClient) GetLeaderApplicationAddress() (string, error) {
	leaderRaftAddress, _ := client.raftNode.LeaderWithID()
	if leaderRaftAddress == "" {
		return "", errors.New("unknown leader")
	}

	transport := &http.Transport{
		Dial: func(network string, address string) (net.Conn, error) {
			return mux.Dial(network, address, 1*time.Second, client.firstByte)
		},
	}
	httpClient := httpUtil.NewClient(&http.Client{
		Transport: transport,
	})

	url := fmt.Sprintf("http://%v%v", leaderRaftAddress, HTTPPath)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	responseBytes, err := httpClient.DoRequest(request)
	if err != nil {
		return "", err
	}
	var response Response
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return "", nil
	}

	// TODO: improve this
	if response.ApplicationAddress[0] == ':' {
		applicationAddress := "http://localhost" + response.ApplicationAddress
		return applicationAddress, nil
	}

	return response.ApplicationAddress, nil
}
