package agent

import (
	"fmt"
	"sync"
	"time"
)

type BackgroundTask struct {
	ID        string
	ToolUseID string
	Command   string
	Status    string
	Result    string
	StartTime time.Time
}

type BackgroundTaskManager struct {
	tasks    map[string]*BackgroundTask
	results  map[string]string
	mu       sync.RWMutex
	wg       sync.WaitGroup
	counter  int64
	handlers map[string]func(map[string]interface{}) string
}

func NewBackgroundTaskManager() *BackgroundTaskManager {
	return &BackgroundTaskManager{
		tasks:    make(map[string]*BackgroundTask),
		results:  make(map[string]string),
		handlers: make(map[string]func(map[string]interface{}) string),
	}
}

func (btm *BackgroundTaskManager) RegisterHandler(toolName string, handler func(map[string]interface{}) string) {
	btm.handlers[toolName] = handler
}

func (btm *BackgroundTaskManager) StartTask(toolName string, toolInput map[string]interface{}, toolUseID string) string {
	btm.mu.Lock()
	defer btm.mu.Unlock()

	btm.counter++
	taskID := fmt.Sprintf("bg_%04d", btm.counter)

	task := &BackgroundTask{
		ID:        taskID,
		ToolUseID: toolUseID,
		Command:   toolInput["command"].(string),
		Status:    "running",
		StartTime: time.Now(),
	}

	btm.tasks[taskID] = task

	btm.wg.Add(1)
	go func() {
		defer btm.wg.Done()

		if handler := btm.handlers[toolName]; handler != nil {
			result := handler(toolInput)
			btm.mu.Lock()
			task.Status = "completed"
			task.Result = result
			btm.results[taskID] = result
			btm.mu.Unlock()

			// Post tool use hooks would go here
		} else {
			btm.mu.Lock()
			task.Status = "failed"
			task.Result = "No handler found for tool: " + toolName
			btm.results[taskID] = task.Result
			btm.mu.Unlock()
		}
	}()

	fmt.Printf("  \033[33m[background] %s: %s\033[0m\n", taskID, truncate(task.Command, 60))
	return taskID
}

func (btm *BackgroundTaskManager) CollectCompletedTasks() []string {
	btm.mu.Lock()
	defer btm.mu.Unlock()

	var notifications []string

	for taskID, task := range btm.tasks {
		if task.Status == "completed" {
			delete(btm.tasks, taskID)
			result := btm.results[taskID]
			delete(btm.results, taskID)

			summary := result
			if len(result) > 200 {
				summary = result[:200] + "..."
			}

			notification := fmt.Sprintf("<task_notification>\n  <task_id>%s</task_id>\n  <status>completed</status>\n  <command>%s</command>\n  <summary>%s</summary>\n</task_notification>",
				taskID, task.Command, summary)

			notifications = append(notifications, notification)
		}
	}

	return notifications
}

func (btm *BackgroundTaskManager) truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
