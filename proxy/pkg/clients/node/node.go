package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	httpUtil "github.com/diegoximenes/distributed_cache/util/pkg/http"
)

type GetResponse struct {
	Value string `json:"value"`
}

type PutInput struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
	TTL   *int64 `json:"ttl"`
}

type NodeClient struct {
	httpClient *httpUtil.HTTPClient
}

func New() *NodeClient {
	httpClient := httpUtil.NewClient(&http.Client{
		Timeout: 2 * time.Second,
	})
	return &NodeClient{
		httpClient: httpClient,
	}
}

func (nodeClient *NodeClient) Get(address string, key string) (*GetResponse, error) {
	url := fmt.Sprintf("http://%v/cache/%v", address, key)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	responseBytes, err := nodeClient.httpClient.DoRequest(request)
	if err != nil {
		return nil, err
	}
	var response GetResponse
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (nodeClient *NodeClient) Put(address string, input *PutInput) error {
	inputJson, err := json.Marshal(input)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%v/cache", address)
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(inputJson))
	if err != nil {
		return err
	}

	_, err = nodeClient.httpClient.DoRequest(request)
	if err != nil {
		return err
	}

	return nil
}
