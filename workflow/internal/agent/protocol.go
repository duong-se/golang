package agent

import (
	"fmt"
	"sync"
	"time"
)

var (
	PendingRequests        = make(map[string]map[string]*ProtocolState)
	pendingRequestsMutex   sync.RWMutex
	ActiveTeammateRequests map[string]bool
	activeMutex            sync.RWMutex
)

const (
	RecoveryStateStruct = "RecoveryState"
)

func registerRequest(state *ProtocolState) {
	pendingRequestsMutex.Lock()
	defer pendingRequestsMutex.Unlock()
	if PendingRequests[state.Sender] == nil {
		PendingRequests[state.Sender] = make(map[string]*ProtocolState)
	}
	PendingRequests[state.Sender][state.Target] = state
	activeMutex.Lock()
	defer activeMutex.Unlock()
	ActiveTeammateRequests[state.Sender+"_"+state.Target] = true
}

func handleResponse(reqID string, approve bool) {
	pendingRequestsMutex.Lock()
	defer pendingRequestsMutex.Unlock()

	for sender := range PendingRequests {
		if stateMap, ok := PendingRequests[sender]; ok {
			if _, ok := stateMap[reqID]; ok {
				delete(stateMap, reqID)
			}
		}
	}
}

func receiveProtocolMessage(sender, target, msgType, payload string) *ProtocolState {
	reqID := fmt.Sprintf("req_%d", time.Now().UnixNano())
	state := &ProtocolState{
		RequestID: reqID,
		Type:      msgType,
		Sender:    sender,
		Target:    target,
		Status:    "pending",
		Payload:   payload,
	}
	registerRequest(state)
	return state
}

func sendProtocolMessage(sender, target, msgType, payload string) *ProtocolState {
	return receiveProtocolMessage(sender, target, msgType, payload)
}

func approveProtocol(reqID string) {
	pendingRequestsMutex.Lock()
	defer pendingRequestsMutex.Unlock()

	for sender := range PendingRequests {
		if stateMap, ok := PendingRequests[sender]; ok {
			if target, ok := stateMap[reqID]; ok {
				target.Status = "approved"
			}
		}
	}
}

func rejectProtocol(reqID string) {
	pendingRequestsMutex.Lock()
	defer pendingRequestsMutex.Unlock()

	for sender := range PendingRequests {
		if stateMap, ok := PendingRequests[sender]; ok {
			if target, ok := stateMap[reqID]; ok {
				target.Status = "rejected"
			}
		}
	}
}

func scanProtocolMessages() []*ProtocolState {
	pendingRequestsMutex.RLock()
	defer pendingRequestsMutex.RUnlock()

	var states []*ProtocolState
	for _, stateMap := range PendingRequests {
		for _, state := range stateMap {
			if state.Status == "pending" {
				states = append(states, state)
			}
		}
	}
	return states
}

func replyToProtocol(target, message string, metadata map[string]interface{}) {
	// Implementation would depend on the actual bus system
	// This is a placeholder
}
