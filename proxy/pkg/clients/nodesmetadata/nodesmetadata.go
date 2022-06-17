package nodesmetadata

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

type NodesMetadataClient struct {
	NodesMetadata NodesMetadata

	httpClient                   http.Client
	nodeMetadataServiceLeaderURL string
}

func New() (*NodesMetadataClient, error) {
	httpClient := http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	nodeMetadataServiceLeaderURL := fmt.Sprintf("%v/nodes", config.Config.NodeMetadataAddress)

	nodesMetadataClient := NodesMetadataClient{
		NodesMetadata: make(NodesMetadata),

		httpClient:                   httpClient,
		nodeMetadataServiceLeaderURL: nodeMetadataServiceLeaderURL,
	}

	err := nodesMetadataClient.sync()
	if err != nil {
		return nil, err
	}

	go nodesMetadataClient.periodicallySync()

	return &nodesMetadataClient, nil
}

func (nodesMetadataClient *NodesMetadataClient) sync() error {
	request, err := http.NewRequest("GET", nodesMetadataClient.nodeMetadataServiceLeaderURL, nil)
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

		leaderURL := location.String()
		nodesMetadataClient.nodeMetadataServiceLeaderURL = leaderURL
		return nodesMetadataClient.sync()
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
	nodesMetadataClient.NodesMetadata = nodesMetadata

	logger.Logger.Info(
		"NodesMetadataClient.sync",
		zap.String("NodesMetadata", fmt.Sprintf("%v", nodesMetadataClient.NodesMetadata)),
	)

	return nil
}

func (nodesMetadataClient *NodesMetadataClient) periodicallySync() {
	for range time.Tick(time.Second * 15) {
		nodesMetadataClient.sync()
	}
}
