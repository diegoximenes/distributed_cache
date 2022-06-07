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
	ID      string `json:"id" binding:"required"`
	Address string `json:"address" binding:"required"`
	Status  string `json:"status" binding:"required"`
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

func writeToFile(obj any, path string) error {
	objJson, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, objJson, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Add(nodeConfigToAdd *NodeConfig) error {
	rwMutex.Lock()
	defer rwMutex.Unlock()

	nodesConfig, err := get()
	if err != nil {
		return err
	}

	nodesConfig[nodeConfigToAdd.ID] = *nodeConfigToAdd

	err = writeToFile(nodesConfig, nodesConfigPath)
	if err != nil {
		return err
	}

	return nil
}

func Delete(id string) error {
	rwMutex.Lock()
	defer rwMutex.Unlock()

	nodesConfig, err := get()
	if err != nil {
		return err
	}

	delete(nodesConfig, id)

	err = writeToFile(nodesConfig, nodesConfigPath)
	if err != nil {
		return err
	}

	return nil
}
