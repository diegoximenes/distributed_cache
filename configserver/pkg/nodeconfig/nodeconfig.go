package nodeconfig

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

const nodesConfigPath = "/tmp/configserver.json"

var rwMutex sync.RWMutex

type NodeConfig struct {
	ID     string `json:"id" binding:"required"`
	IP     string `json:"ip" binding:"required"`
	Status string `json:"status" binding:"required"`
}

func get() (map[string]NodeConfig, error) {
	jsonFile, err := os.Open(nodesConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]NodeConfig{}, nil
		}
		return map[string]NodeConfig{}, err
	}

	bytesFile, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return map[string]NodeConfig{}, err
	}

	var nodesConfig map[string]NodeConfig
	err = json.Unmarshal(bytesFile, &nodesConfig)
	if err != nil {
		return map[string]NodeConfig{}, err
	}

	return nodesConfig, nil
}

func Get() (map[string]NodeConfig, error) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	return get()
}

func Add(nodeConfigToAdd *NodeConfig) error {
	rwMutex.Lock()
	defer rwMutex.Unlock()

	nodesConfig, err := get()
	if err != nil {
		return err
	}

	nodesConfig[nodeConfigToAdd.ID] = *nodeConfigToAdd

	nodesConfigJson, err := json.Marshal(nodesConfig)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(nodesConfigPath, nodesConfigJson, 0644)
	if err != nil {
		return err
	}

	return nil
}
