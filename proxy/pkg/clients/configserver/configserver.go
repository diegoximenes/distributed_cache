package configserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	httpUtil "github.com/diegoximenes/distributed_key_value_cache/proxy/pkg/util/http"
)

const configServerURL = "http://localhost:28000/node"

type NodeConfig struct {
	ID      string `json:"id" binding:"required"`
	Address string `json:"address" binding:"required"`
	Status  string `json:"status" binding:"required"`
}

type NodesConfig map[string]NodeConfig

type ConfigServerClient struct {
	NodesConfig *NodesConfig
}

func New() (*ConfigServerClient, error) {
	nodesConfig := make(NodesConfig)
	configserverClient := &ConfigServerClient{
		NodesConfig: &nodesConfig,
	}

	err := configserverClient.get()
	if err != nil {
		return nil, err
	}

	go configserverClient.periodicallySync()

	return configserverClient, nil
}

func (configserverClient *ConfigServerClient) get() error {
	request, err := http.NewRequest("GET", configServerURL, nil)
	if err != nil {
		return err
	}

	responseBytes, err := httpUtil.DoRequest(request)
	if err != nil {
		return err
	}
	var response NodesConfig
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return err
	}
	configserverClient.NodesConfig = &response

	fmt.Println(response)

	return nil
}

func (configserverClient *ConfigServerClient) periodicallySync() {
	for range time.Tick(time.Second * 15) {
		configserverClient.get()
	}
}
