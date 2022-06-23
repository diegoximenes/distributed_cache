package metadata

import (
	"github.com/diegoximenes/distributed_cache/nodesmetadata/internal/config"
	"github.com/diegoximenes/distributed_cache/nodesmetadata/pkg/net/sse"
	"github.com/hashicorp/raft"
)

type RaftEvent struct {
	Type  string      `json:"type"`
	Event interface{} `json:"event"`
}

func observationType(observation *raft.Observation) string {
	_, isRequestVoteRequest := observation.Data.(raft.RequestVoteRequest)
	if isRequestVoteRequest {
		return "RequestVoteRequest"
	}

	_, isRaftState := observation.Data.(raft.RaftState)
	if isRaftState {
		return "RaftState"
	}

	_, isPeerObservation := observation.Data.(raft.PeerObservation)
	if isPeerObservation {
		return "PeerObservation"
	}

	_, isLeaderObservation := observation.Data.(raft.LeaderObservation)
	if isLeaderObservation {
		return "LeaderObservation"
	}

	return ""
}

func filterObservation(observation *raft.Observation) bool {
	_, isRequestVoteRequest := observation.Data.(raft.RequestVoteRequest)
	_, isRaftState := observation.Data.(raft.RaftState)
	if isRequestVoteRequest || isRaftState {
		return false
	}
	return true
}

func NewSSE(raftNode *raft.Raft, nodesSSE *sse.SSE) *sse.SSE {
	raftMetadataSSE := sse.New()

	observationChan := make(chan raft.Observation)
	observer := raft.NewObserver(observationChan, false, filterObservation)
	go func() {
		for {
			observation := <-observationChan

			leaderObservation, isLeaderObservation := observation.Data.(raft.LeaderObservation)
			if isLeaderObservation && (leaderObservation.LeaderID != raft.ServerID(config.Config.RaftId)) {
				raftMetadataSSE.CloseAllClientsChan <- true
				nodesSSE.CloseAllClientsChan <- true
			} else {
				raftEvent := RaftEvent{
					Type:  observationType(&observation),
					Event: observation.Data,
				}
				raftMetadataSSE.EventsToSend <- &raftEvent
			}
		}
	}()
	raftNode.RegisterObserver(observer)

	return raftMetadataSSE
}
