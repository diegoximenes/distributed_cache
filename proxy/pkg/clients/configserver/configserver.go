package configserver

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

type NodeConfig struct {
	ID      string `json:"id" binding:"required"`
	Address string `json:"address" binding:"required"`
	Status  string `json:"status" binding:"required"`
}

type NodesConfig map[string]NodeConfig

type ConfigServerClient struct {
	NodesConfig NodesConfig
}

func New() (*ConfigServerClient, error) {
	nodesConfig := make(NodesConfig)
	configserverClient := ConfigServerClient{
		NodesConfig: nodesConfig,
	}

	err := configserverClient.sync()
	if err != nil {
		return nil, err
	}

	go configserverClient.periodicallySync()

	return &configserverClient, nil
}

func (configServerClient *ConfigServerClient) sync() error {
	url := fmt.Sprintf("%s/node", config.Config.ConfigServerAddress)
	request, err := http.NewRequest("GET", url, nil)
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
	configServerClient.NodesConfig = response

	logger.Logger.Info(
		"ConfigServerClient.sync",
		zap.String("NodesConfig", fmt.Sprintf("%v", configServerClient.NodesConfig)),
	)

	return nil
}

func (configServerClient *ConfigServerClient) periodicallySync() {
	for range time.Tick(time.Second * 15) {
		configServerClient.sync()
	}
}
