package nodesmetadata

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/diegoximenes/distributed_cache/proxy/internal/config"
	"github.com/diegoximenes/distributed_cache/proxy/internal/keypartition"
	"github.com/diegoximenes/distributed_cache/proxy/internal/logger"
	"go.uber.org/zap"
)

type NodeMetadata struct {
	ID      string `json:"id" binding:"required"`
	Address string `json:"address" binding:"required"`
}

type NodesMetadata map[string]NodeMetadata

type raftNodeMetadata struct {
	NodeID             string `json:"nodeID"`
	ApplicationAddress string `json:"applicationAddress"`
	RaftAddress        string `json:"raftAddress"`
}

type raftNodesMetadata map[string]*raftNodeMetadata

type raftMetadata struct {
	NodesMetadata raftNodesMetadata `json:"nodesMetadata"`
	LeaderNodeID  string            `json:"leaderNodeID"`
}

type NodesMetadataClient struct {
	NodesMetadata NodesMetadata

	httpClient               http.Client
	httpClientSSE            http.Client
	leaderApplicationAddress string
	raftMetadata             raftMetadata
	keyPartitionStrategy     keypartition.KeyPartitionStrategy
	// is used to keep NodesMetadata and keyPartitionStrategy.nodesID in a consistent way
	syncNodesMetadataMutex sync.Mutex
}

const (
	nodesMetadataPath    = "/nodes"
	nodesMetadataSSEPath = "/nodes/sse"
	raftMetadataPath     = "/raft/metadata"
	raftMetadataSSEPath  = "/raft/metadata/sse"
)

var (
	sseDataRegexp, _ = regexp.Compile("^data:.*\n$")
)

func New(
	keyPartitionStrategy keypartition.KeyPartitionStrategy,
) *NodesMetadataClient {
	httpClient := http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 2 * time.Second,
	}
	httpClientSSE := http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	leaderApplicationAddress := config.Config.NodesMetadataAddress

	nodesMetadataClient := NodesMetadataClient{
		NodesMetadata: make(NodesMetadata),

		httpClient:               httpClient,
		httpClientSSE:            httpClientSSE,
		leaderApplicationAddress: leaderApplicationAddress,
		raftMetadata:             raftMetadata{},
		keyPartitionStrategy:     keyPartitionStrategy,
	}

	go nodesMetadataClient.syncRaftMetadataSSE()
	go nodesMetadataClient.syncNodesMetadataSSE()

	go nodesMetadataClient.periodicallySync()

	return &nodesMetadataClient
}

func (nodesMetadataClient *NodesMetadataClient) getAddressToUse(
	addressesTried map[string]bool,
) (string, error) {
	if _, exists := addressesTried[nodesMetadataClient.leaderApplicationAddress]; !exists {
		return nodesMetadataClient.leaderApplicationAddress, nil
	}
	for _, raftNodeMetadata := range nodesMetadataClient.raftMetadata.NodesMetadata {
		if raftNodeMetadata.ApplicationAddress == "" {
			continue
		}
		if _, exists := addressesTried[raftNodeMetadata.ApplicationAddress]; !exists {
			return raftNodeMetadata.ApplicationAddress, nil
		}
	}
	return "", errors.New("no address available")
}

func (nodesMetadataClient *NodesMetadataClient) sync(
	httpClient *http.Client,
	urlPath string,
	stateUpdater func(io.ReadCloser) error,
	addressesTried map[string]bool,
) error {
	address, err := nodesMetadataClient.getAddressToUse(addressesTried)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%v%v", address, urlPath)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		addressesTried[address] = true
		return nodesMetadataClient.sync(httpClient, urlPath, stateUpdater, addressesTried)
	}
	response, err := httpClient.Do(request)
	if err != nil {
		addressesTried[address] = true
		return nodesMetadataClient.sync(httpClient, urlPath, stateUpdater, addressesTried)
	}
	defer response.Body.Close()

	if (response.StatusCode >= 300) && (response.StatusCode < 400) {
		// handle redirect to leader

		location, err := response.Location()
		if err != nil {
			return err
		}

		leaderAddress := strings.Split(location.String(), urlPath)[0]
		leaderAddress = strings.Split(leaderAddress, "http://")[1]
		nodesMetadataClient.leaderApplicationAddress = leaderAddress
	}
	if (response.StatusCode < 200) || (response.StatusCode >= 300) {
		addressesTried[address] = true
		return nodesMetadataClient.sync(httpClient, urlPath, stateUpdater, addressesTried)
	}

	return stateUpdater(response.Body)
}

func (nodesMetadataClient *NodesMetadataClient) nodesMetadataStateUpdater(
	body io.ReadCloser,
) error {
	responseBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	var nodesMetadata NodesMetadata
	err = json.Unmarshal(responseBytes, &nodesMetadata)
	if err != nil {
		return err
	}
	nodesMetadataClient.NodesMetadata = nodesMetadata

	nodesID := make([]string, len(nodesMetadata))
	i := 0
	for nodeID := range nodesMetadata {
		nodesID[i] = nodeID
		i++
	}
	nodesMetadataClient.keyPartitionStrategy.UpdateNodes(nodesID)

	logger.Logger.Info(
		"NodesMetadataClient.nodesMetadataStateUpdater",
		zap.String("NodesMetadata", fmt.Sprintf("%v", nodesMetadataClient.NodesMetadata)),
	)

	return nil
}

func (nodesMetadataClient *NodesMetadataClient) raftMetadataStateUpdater(
	body io.ReadCloser,
) error {
	responseBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	var raftMetadata raftMetadata
	err = json.Unmarshal(responseBytes, &raftMetadata)
	if err != nil {
		return err
	}
	nodesMetadataClient.raftMetadata = raftMetadata

	if (raftMetadata.LeaderNodeID != "") &&
		(raftMetadata.NodesMetadata[raftMetadata.LeaderNodeID].ApplicationAddress != "") {
		nodesMetadataClient.leaderApplicationAddress =
			raftMetadata.NodesMetadata[raftMetadata.LeaderNodeID].ApplicationAddress
	}

	logger.Logger.Info(
		"NodesMetadataClient.raftMetadataStateUpdater",
		zap.String("raftMetadata", fmt.Sprintf("%v", nodesMetadataClient.raftMetadata)),
	)

	return nil
}

func (nodesMetadataClient *NodesMetadataClient) syncNodesMetadata() error {
	nodesMetadataClient.syncNodesMetadataMutex.Lock()
	defer nodesMetadataClient.syncNodesMetadataMutex.Unlock()

	return nodesMetadataClient.sync(
		&nodesMetadataClient.httpClient,
		nodesMetadataPath,
		nodesMetadataClient.nodesMetadataStateUpdater,
		make(map[string]bool),
	)
}

func (nodesMetadataClient *NodesMetadataClient) syncRaftMetadata() error {
	return nodesMetadataClient.sync(
		&nodesMetadataClient.httpClient,
		raftMetadataPath,
		nodesMetadataClient.raftMetadataStateUpdater,
		make(map[string]bool),
	)
}

func (nodesMetadataClient *NodesMetadataClient) sseStateUpdater(
	sync func() error,
) func(io.ReadCloser) error {
	return func(body io.ReadCloser) error {
		reader := bufio.NewReader(body)
		sync()
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return err
			}
			if sseDataRegexp.MatchString(string(line)) {
				sync()
			}
		}
	}
}

func (nodesMetadataClient *NodesMetadataClient) syncRaftMetadataSSE() {
	for {
		err := nodesMetadataClient.sync(
			&nodesMetadataClient.httpClientSSE,
			raftMetadataSSEPath,
			nodesMetadataClient.sseStateUpdater(nodesMetadataClient.syncRaftMetadata),
			make(map[string]bool),
		)
		logger.Logger.Warn(err.Error())
	}
}

func (nodesMetadataClient *NodesMetadataClient) syncNodesMetadataSSE() {
	for {
		err := nodesMetadataClient.sync(
			&nodesMetadataClient.httpClientSSE,
			nodesMetadataSSEPath,
			nodesMetadataClient.sseStateUpdater(nodesMetadataClient.syncNodesMetadata),
			make(map[string]bool),
		)
		logger.Logger.Warn(err.Error())
	}
}

func (nodesMetadataClient *NodesMetadataClient) periodicallySync() {
	for range time.Tick(time.Minute * 1) {
		if err := nodesMetadataClient.syncRaftMetadata(); err != nil {
			logger.Logger.Warn(err.Error())
		}
		if err := nodesMetadataClient.syncNodesMetadata(); err != nil {
			logger.Logger.Warn(err.Error())
		}
	}
}
