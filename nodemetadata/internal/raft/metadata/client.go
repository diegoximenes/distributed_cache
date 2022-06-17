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

func (client *RaftNodeMetadataClient) getApplicationAddress(raftAddress string) (string, error) {
	transport := &http.Transport{
		Dial: func(network string, address string) (net.Conn, error) {
			return mux.Dial(network, address, 1*time.Second, client.firstByte)
		},
	}
	httpClient := httpUtil.NewClient(&http.Client{
		Transport: transport,
	})

	url := fmt.Sprintf("http://%v%v", raftAddress, HTTPPath)
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

func (client *RaftNodeMetadataClient) GetLeaderApplicationAddress() (string, error) {
	leaderRaftAddress, _ := client.raftNode.LeaderWithID()
	if leaderRaftAddress == "" {
		return "", errors.New("unknown leader")
	}
	return client.getApplicationAddress(string(leaderRaftAddress))
}

func (client *RaftNodeMetadataClient) addApplicationAddressToChannel(raftAddress string, applicationAddressesChannel chan string) {
	applicationAddress, err := client.getApplicationAddress(raftAddress)
	if err == nil {
		applicationAddressesChannel <- applicationAddress
	} else {
		applicationAddressesChannel <- ""
	}
}

func (client *RaftNodeMetadataClient) GetNodesApplicationAddresses() []string {
	nodes := client.raftNode.GetConfiguration().Configuration().Servers

	applicationAddressesChannel := make(chan string, len(nodes))
	defer close(applicationAddressesChannel)
	for _, node := range nodes {
		go client.addApplicationAddressToChannel(string(node.Address), applicationAddressesChannel)
	}

	var applicationAddresses []string
	for i := 0; i < len(nodes); i++ {
		applicationAddress := <-applicationAddressesChannel
		if applicationAddress != "" {
			applicationAddresses = append(applicationAddresses, applicationAddress)
		}
	}

	return applicationAddresses
}
