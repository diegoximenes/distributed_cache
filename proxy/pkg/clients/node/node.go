package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	httpUtil "github.com/diegoximenes/distributed_cache/proxy/pkg/util/http"
)

type GetResponse struct {
	Value string `json:"value"`
}

type PutInput struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
	TTL   *int64 `json:"ttl"`
}

func Get(address string, key string) (*GetResponse, error) {
	url := fmt.Sprintf("%v/cache/%v", address, key)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	responseBytes, err := httpUtil.DoRequest(request)
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

func Put(address string, input *PutInput) error {
	inputJson, err := json.Marshal(input)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%v/cache", address)
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(inputJson))
	if err != nil {
		return err
	}

	_, err = httpUtil.DoRequest(request)
	if err != nil {
		return err
	}

	return nil
}
