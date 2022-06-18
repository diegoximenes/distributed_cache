package nodesmetadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/diegoximenes/distributed_cache/proxy/internal/config"
	"github.com/diegoximenes/distributed_cache/proxy/internal/util/logger"
	"go.uber.org/zap"
)

type NodeMetadata struct {
	ID      string `json:"id" binding:"required"`
	Address string `json:"address" binding:"required"`
}

type NodesMetadata map[string]NodeMetadata

type raftNodeMetadata struct {
	NodeID             string `json:"nodeID"`
	ApplicationAddress string `json:"applicationAddress"`
	RaftAddress        string `json:"raftAddress"`
}

type raftNodesMetadata map[string]*raftNodeMetadata

type raftMetadata struct {
	NodesMetadata raftNodesMetadata `json:"nodesMetadata"`
	LeaderNodeID  string            `json:"leaderNodeID"`
}

type NodesMetadataClient struct {
	NodesMetadata NodesMetadata

	httpClient               http.Client
	leaderApplicationAddress string
	raftMetadata             raftMetadata
}

const (
	nodesMetadataPath = "/nodes"
	raftMetadataPath  = "/raft/metadata"
)

func New() (*NodesMetadataClient, error) {
	httpClient := http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 2 * time.Second,
	}

	leaderApplicationAddress := config.Config.NodesMetadataAddress

	nodesMetadataClient := NodesMetadataClient{
		NodesMetadata: make(NodesMetadata),

		httpClient:               httpClient,
		leaderApplicationAddress: leaderApplicationAddress,
		raftMetadata:             raftMetadata{},
	}

	err := nodesMetadataClient.syncRaftMetadata()
	if err != nil {
		return nil, err
	}

	err = nodesMetadataClient.syncNodesMetadata()
	if err != nil {
		return nil, err
	}

	go nodesMetadataClient.periodicallySync()

	return &nodesMetadataClient, nil
}

func (nodesMetadataClient *NodesMetadataClient) getAddressToUse(
	addressesTries map[string]bool,
) (string, error) {
	if _, exists := addressesTries[nodesMetadataClient.leaderApplicationAddress]; !exists {
		return nodesMetadataClient.leaderApplicationAddress, nil
	}
	for _, raftNodeMetadata := range nodesMetadataClient.raftMetadata.NodesMetadata {
		if raftNodeMetadata.ApplicationAddress == "" {
			continue
		}
		if _, exists := addressesTries[raftNodeMetadata.ApplicationAddress]; !exists {
			return raftNodeMetadata.ApplicationAddress, nil
		}
	}
	return "", errors.New("no address available")
}

func (nodesMetadataClient *NodesMetadataClient) sync(
	urlPath string,
	stateUpdater func(responseBytes []byte) error,
	addressesTried map[string]bool,
) error {
	address, err := nodesMetadataClient.getAddressToUse(addressesTried)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%v%v", address, urlPath)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	response, err := nodesMetadataClient.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if (response.StatusCode >= 300) && (response.StatusCode < 400) {
		location, err := response.Location()
		if err != nil {
			return err
		}

		leaderAddress := strings.Split(location.String(), urlPath)[0]
		nodesMetadataClient.leaderApplicationAddress = leaderAddress
	}
	if (response.StatusCode < 200) || (response.StatusCode >= 300) {
		addressesTried[address] = true
		return nodesMetadataClient.sync(urlPath, stateUpdater, addressesTried)
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return stateUpdater(responseBytes)
}

func (nodesMetadataClient *NodesMetadataClient) nodesMetadataStateUpdater(
	responseBytes []byte,
) error {
	var nodesMetadata NodesMetadata
	err := json.Unmarshal(responseBytes, &nodesMetadata)
	if err != nil {
		return err
	}
	nodesMetadataClient.NodesMetadata = nodesMetadata

	logger.Logger.Info(
		"NodesMetadataClient.nodesMetadataStateUpdater",
		zap.String("NodesMetadata", fmt.Sprintf("%v", nodesMetadataClient.NodesMetadata)),
	)

	return nil
}

func (nodesMetadataClient *NodesMetadataClient) raftMetadataStateUpdater(
	responseBytes []byte,
) error {
	var raftMetadata raftMetadata
	err := json.Unmarshal(responseBytes, &raftMetadata)
	if err != nil {
		return err
	}
	nodesMetadataClient.raftMetadata = raftMetadata

	if (raftMetadata.LeaderNodeID != "") &&
		(raftMetadata.NodesMetadata[raftMetadata.LeaderNodeID].ApplicationAddress != "") {
		nodesMetadataClient.leaderApplicationAddress =
			raftMetadata.NodesMetadata[raftMetadata.LeaderNodeID].ApplicationAddress
	}

	logger.Logger.Info(
		"NodesMetadataClient.raftMetadataStateUpdater",
		zap.String("raftMetadata",
			fmt.Sprintf("%v", nodesMetadataClient.raftMetadata)),
	)

	return nil
}

func (nodesMetadataClient *NodesMetadataClient) syncNodesMetadata() error {
	return nodesMetadataClient.sync(
		nodesMetadataPath,
		nodesMetadataClient.nodesMetadataStateUpdater,
		make(map[string]bool),
	)
}

func (nodesMetadataClient *NodesMetadataClient) syncRaftMetadata() error {
	return nodesMetadataClient.sync(
		raftMetadataPath,
		nodesMetadataClient.raftMetadataStateUpdater,
		make(map[string]bool),
	)
}

func (nodesMetadataClient *NodesMetadataClient) periodicallySync() {
	for range time.Tick(time.Second * 15) {
		nodesMetadataClient.syncRaftMetadata()
		nodesMetadataClient.syncNodesMetadata()
	}
}
