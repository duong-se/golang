package agent

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type Agent struct {
	name     string
	role     string
	messages []Message
	bus      *MessageBus
	tasks    *TaskManager
	config   *SystemConfig
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

type AgentConfig struct {
	Model         string
	MaxTokens     int
	PrimaryModel  string
	FallbackModel string
}

func NewAgent(name, role string, config *AgentConfig) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		name:     name,
		role:     role,
		messages: make([]Message, 0),
		bus:      NewMessageBus(),
		tasks:    NewTaskManager(),
		config:   &SystemConfig{WorkDir: WorkDir},
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (a *Agent) Start() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		a.cancel()
	}()

	return a.runLoop()
}

func (a *Agent) runLoop() error {
	ticker := time.NewTicker(IDLE_POLL_INTERVAL * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return a.ctx.Err()

		case <-ticker.C:
			a.processInbox()
			a.processUnclaimedTasks()

		default:
			a.processInbox()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (a *Agent) processInbox() {
	msgs := a.bus.ReadInbox(a.name)
	for _, msg := range msgs {
		if msg.MsgType == "shutdown_request" {
			a.handleShutdown(msg)
			return
		}
		a.handleMessage(msg)
	}
}

func (a *Agent) processUnclaimedTasks() {
	unclaimed := a.tasks.ScanUnclaimed()
	if len(unclaimed) == 0 {
		return
	}

	task := unclaimed[0]
	result, err := a.tasks.Claim(task.ID, a.name)
	if err != nil {
		fmt.Printf("Failed to claim task %s: %v\n", task.ID, err)
		return
	}

	fmt.Printf("  \033[36m[claim] %s → in_progress\033[0m\n", task.Subject)

	wtInfo := ""
	if task.Worktree != "" {
		wtPath := filepath.Join(WorktreeDir, task.Worktree)
		wtInfo = fmt.Sprintf("\nWork directory: %s", wtPath)
	}

	a.messages = append(a.messages, Message{
		Role:    "user",
		Content: fmt.Sprintf("<auto-claimed>Task %s: %s%s</auto-claimed>", task.ID, task.Subject, wtInfo),
	})
}

func (a *Agent) handleMessage(msg Message) {
	a.messages = append(a.messages, Message{
		Role:    "user",
		Content: "<inbox>" + msg.Content + "</inbox>",
	})
}

func (a *Agent) handleShutdown(msg Message) {
	if reqID, ok := msg.Metadata["request_id"].(string); ok {
		a.bus.Send(a.name, "lead", "Shutting down.", "shutdown_response", map[string]interface{}{
			"request_id": reqID,
			"approve":    true,
		})
	}
	a.cancel()
}

func (a *Agent) SpawnTeammate(name, role, prompt string) string {
	if _, exists := ActiveTeammateRequests[name]; exists {
		return fmt.Sprintf("Teammate '%s' already exists", name)
	}

	teammate := &Teammate{
		Name:      name,
		Role:      role,
		Context:   a.ctx,
		completed: make(chan struct{}),
	}

	go teammate.Run(func(ctx context.Context) error {
		return a.runTeammate(ctx, name, role, prompt)
	})

	ActiveTeammateRequests[name] = true
	return fmt.Sprintf("Teammate '%s' spawned as %s", name, role)
}

func (a *Agent) runTeammate(ctx context.Context, name, role, prompt string) error {
	// Teammate execution logic
	return nil
}
