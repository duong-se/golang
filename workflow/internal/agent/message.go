package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Message struct {
	FromAgent string                 `json:"from"`
	ToAgent   string                 `json:"to"`
	Content   string                 `json:"content"`
	MsgType   string                 `json:"type"`
	Timestamp float64                `json:"ts"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type MessageBus struct {
	mu      sync.RWMutex
	dirPath string
}

func NewMessageBus() *MessageBus {
	if mailboxDir == "" {
		mailboxDir = filepath.Join(WorkDir, ".mailboxes")
	}
	os.MkdirAll(mailboxDir, 0755)
	return &MessageBus{dirPath: mailboxDir}
}

func (mb *MessageBus) Send(fromAgent, toAgent, content, msgType string, metadata map[string]interface{}) {
	msg := Message{
		FromAgent: fromAgent,
		ToAgent:   toAgent,
		Content:   content,
		MsgType:   msgType,
		Timestamp: time.Now().Unix(),
		Metadata:  metadata,
	}

	msgData, _ := json.Marshal(msg)
	mb.mu.Lock()
	destPath := filepath.Join(mb.dirPath, toAgent+".jsonl")
	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		mb.mu.Unlock()
		return
	}
	fmt.Fprintln(f, string(msgData))
	f.Close()
	mb.mu.Unlock()

	fmt.Printf("  \033[33m[bus] %s → %s: (%s) %s\033[0m\n", fromAgent, toAgent, msgType, content[:50])
}

func (mb *MessageBus) ReadInbox(agent string) []Message {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	destPath := filepath.Join(mb.dirPath, agent+".jsonl")

	data, err := ioutil.ReadFile(destPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Message{}
		}
		return []Message{}
	}

	var msgs []Message
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var msg Message
		if err := json.Unmarshal([]byte(line), &msg); err == nil {
			msgs = append(msgs, msg)
		}
	}

	if len(msgs) > 0 {
		_ = os.Remove(destPath)
	}

	return msgs
}
