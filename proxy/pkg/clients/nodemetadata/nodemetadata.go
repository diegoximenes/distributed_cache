package nodemetadata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/diegoximenes/distributed_key_value_cache/proxy/internal/config"
	"github.com/diegoximenes/distributed_key_value_cache/proxy/internal/util/logger"
	httpUtil "github.com/diegoximenes/distributed_key_value_cache/proxy/pkg/util/http"
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
}

func New() (*NodeMetadataClient, error) {
	nodesMetadata := make(NodesMetadata)
	nodeMetadataClient := NodeMetadataClient{
		NodesMetadata: nodesMetadata,
	}

	err := nodeMetadataClient.sync()
	if err != nil {
		return nil, err
	}

	go nodeMetadataClient.periodicallySync()

	return &nodeMetadataClient, nil
}

func (nodeMetadataClient *NodeMetadataClient) sync() error {
	url := fmt.Sprintf("%s/node", config.Config.NodeMetadataAddress)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	responseBytes, err := httpUtil.DoRequest(request)
	if err != nil {
		return err
	}
	var response NodesMetadata
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return err
	}
	nodeMetadataClient.NodesMetadata = response

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
