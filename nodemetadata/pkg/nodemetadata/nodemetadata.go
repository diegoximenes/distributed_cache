package nodemetadata

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

const nodesMetadataPath = "/tmp/nodemetadata.json"

var rwMutex sync.RWMutex

type NodeMetadata struct {
	ID      string `json:"id" binding:"required"`
	Address string `json:"address" binding:"required"`
	Status  string `json:"status" binding:"required"`
}

func get() (map[string]NodeMetadata, error) {
	jsonFile, err := os.Open(nodesMetadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]NodeMetadata{}, nil
		}
		return map[string]NodeMetadata{}, err
	}

	bytesFile, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return map[string]NodeMetadata{}, err
	}

	var nodesMetadata map[string]NodeMetadata
	err = json.Unmarshal(bytesFile, &nodesMetadata)
	if err != nil {
		return map[string]NodeMetadata{}, err
	}

	return nodesMetadata, nil
}

func Get() (map[string]NodeMetadata, error) {
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

func Add(nodeMetadataToAdd *NodeMetadata) error {
	rwMutex.Lock()
	defer rwMutex.Unlock()

	nodesMetadata, err := get()
	if err != nil {
		return err
	}

	nodesMetadata[nodeMetadataToAdd.ID] = *nodeMetadataToAdd

	err = writeToFile(nodesMetadata, nodesMetadataPath)
	if err != nil {
		return err
	}

	return nil
}

func Delete(id string) error {
	rwMutex.Lock()
	defer rwMutex.Unlock()

	nodesMetadata, err := get()
	if err != nil {
		return err
	}

	delete(nodesMetadata, id)

	err = writeToFile(nodesMetadata, nodesMetadataPath)
	if err != nil {
		return err
	}

	return nil
}
