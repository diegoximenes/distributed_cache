package nodemetadata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

type NodeMetadataClient struct {
	NodesMetadata NodesMetadata

	httpClient                   http.Client
	nodeMetadataServiceLeaderURL string
}

func New() (*NodeMetadataClient, error) {
	httpClient := http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	nodeMetadataServiceLeaderURL := fmt.Sprintf("%v/node", config.Config.NodeMetadataAddress)

	nodeMetadataClient := NodeMetadataClient{
		NodesMetadata: make(NodesMetadata),

		httpClient:                   httpClient,
		nodeMetadataServiceLeaderURL: nodeMetadataServiceLeaderURL,
	}

	err := nodeMetadataClient.sync()
	if err != nil {
		return nil, err
	}

	go nodeMetadataClient.periodicallySync()

	return &nodeMetadataClient, nil
}

func (nodeMetadataClient *NodeMetadataClient) sync() error {
	request, err := http.NewRequest("GET", nodeMetadataClient.nodeMetadataServiceLeaderURL, nil)
	if err != nil {
		return err
	}
	response, err := nodeMetadataClient.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if (response.StatusCode >= 300) && (response.StatusCode < 400) {
		location, err := response.Location()
		if err != nil {
			return err
		}

		leaderURL := location.String()
		nodeMetadataClient.nodeMetadataServiceLeaderURL = leaderURL
		return nodeMetadataClient.sync()
	} else if (response.StatusCode < 200) || (response.StatusCode >= 400) {
		// TODO: get all raft nodes addresses
		return err
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	var nodesMetadata NodesMetadata
	err = json.Unmarshal(responseBytes, &nodesMetadata)
	if err != nil {
		return err
	}
	nodeMetadataClient.NodesMetadata = nodesMetadata

	logger.Logger.Info(
		"NodeMetadataClient.sync",
		zap.String("NodesMetadata", fmt.Sprintf("%v", nodeMetadataClient.NodesMetadata)),
	)

	return nil
}

func (nodeMetadataClient *NodeMetadataClient) periodicallySync() {
	for range time.Tick(time.Second * 15) {
		nodeMetadataClient.sync()
	}
}
