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
	Status  string `json:"status" binding:"required"`
}

type NodesMetadata map[string]NodeMetadata

type RaftNodesMetadata struct {
	NodesApplicationAddresses []string `json:"nodesApplicationAddresses"`
}

type NodesMetadataClient struct {
	NodesMetadata NodesMetadata

	httpClient                        http.Client
	nodesMetadataServiceLeaderAddress string
	nodesMetadataServiceAddresses     []string
}

const (
	nodesMetadataPath     = "/nodes"
	raftNodesMetadataPath = "/raft/nodes"
)

func New() (*NodesMetadataClient, error) {
	httpClient := http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 2 * time.Second,
	}

	nodesMetadataServiceLeaderAddress := config.Config.NodesMetadataAddress

	nodesMetadataClient := NodesMetadataClient{
		NodesMetadata: make(NodesMetadata),

		httpClient:                        httpClient,
		nodesMetadataServiceLeaderAddress: nodesMetadataServiceLeaderAddress,
		nodesMetadataServiceAddresses:     make([]string, 0),
	}

	err := nodesMetadataClient.syncRaftNodesMetadata()
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

func (nodesMetadataClient *NodesMetadataClient) getAddressToUse(addressesTries map[string]bool) (string, error) {
	if _, exists := addressesTries[nodesMetadataClient.nodesMetadataServiceLeaderAddress]; !exists {
		return nodesMetadataClient.nodesMetadataServiceLeaderAddress, nil
	}
	for _, address := range nodesMetadataClient.nodesMetadataServiceAddresses {
		if _, exists := addressesTries[address]; !exists {
			return address, nil
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
		nodesMetadataClient.nodesMetadataServiceLeaderAddress = leaderAddress
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

func (nodesMetadataClient *NodesMetadataClient) nodesMetadataStateUpdater(responseBytes []byte) error {
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

func (nodesMetadataClient *NodesMetadataClient) raftNodesMetadataStateUpdater(responseBytes []byte) error {
	var raftNodesMetadata RaftNodesMetadata
	err := json.Unmarshal(responseBytes, &raftNodesMetadata)
	if err != nil {
		return err
	}
	nodesMetadataClient.nodesMetadataServiceAddresses = raftNodesMetadata.NodesApplicationAddresses

	logger.Logger.Info(
		"NodesMetadataClient.raftNodesMetadataStateUpdater",
		zap.String("nodesMetadataServiceAddresses", fmt.Sprintf("%v", nodesMetadataClient.nodesMetadataServiceAddresses)),
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

func (nodesMetadataClient *NodesMetadataClient) syncRaftNodesMetadata() error {
	return nodesMetadataClient.sync(
		raftNodesMetadataPath,
		nodesMetadataClient.raftNodesMetadataStateUpdater,
		make(map[string]bool),
	)
}

func (nodesMetadataClient *NodesMetadataClient) periodicallySync() {
	for range time.Tick(time.Second * 15) {
		nodesMetadataClient.syncRaftNodesMetadata()
		nodesMetadataClient.syncNodesMetadata()
	}
}
