package agent

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Task represents a task in the system.
type Task struct {
	ID          string   `json:"id"`
	Subject     string   `json:"subject"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Owner       *string  `json:"owner,omitempty"`
	BlockedBy   []string `json:"blockedBy,omitempty"`
	Worktree    *string  `json:"worktree,omitempty"`
}

// TasksDir is the directory where task files are stored.
var TasksDir string

// CurrentTodos holds the current TODO list.
var CurrentTodos []map[string]interface{}

func init() {
	// Set TasksDir to the current working directory + "/.tasks"
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	TasksDir = filepath.Join(wd, ".tasks")
	// Ensure the directory exists.
	if err := os.MkdirAll(TasksDir, 0755); err != nil {
		panic(err)
	}
}

// taskPath returns the file path for a given task ID.
func taskPath(taskID string) string {
	return filepath.Join(TasksDir, taskID+".json")
}

// CreateTask creates a new task and saves it to disk.
func CreateTask(subject, description string, blockedBy []string) *Task {
	task := &Task{
		ID:          fmt.Sprintf("task_%d_%04d", time.Now().Unix(), rand.Intn(10000)),
		Subject:     subject,
		Description: description,
		Status:      "pending",
		Owner:       nil,
		BlockedBy:   blockedBy,
		Worktree:    nil,
	}
	if err := saveTask(task); err != nil {
		panic(err)
	}
	return task
}

// saveTask marshals the task to JSON and writes it to disk.
func saveTask(task *Task) error {
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(taskPath(task.ID), data, 0644)
}

// loadTask reads a task from disk by ID.
func loadTask(taskID string) (*Task, error) {
	data, err := os.ReadFile(taskPath(taskID))
	if err != nil {
		return nil, err
	}
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks returns all tasks sorted by ID.
func ListTasks() ([]*Task, error) {
	files, err := filepath.Glob(filepath.Join(TasksDir, "task_*.json"))
	if err != nil {
		return nil, err
	}
	var tasks []*Task
	for _, file := range files {
		task, err := loadTask(filepath.Base(file)[:len(filepath.Base(file))-5])
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return tasks, nil
}

// GetTaskJSON returns the JSON representation of a task.
func GetTaskJSON(taskID string) (string, error) {
	task, err := loadTask(taskID)
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// canStart checks if a task can be started based on its blockers.
func CanStart(taskID string) bool {
	task, err := loadTask(taskID)
	if err != nil {
		return false
	}
	for _, depID := range task.BlockedBy {
		if _, err := os.Stat(taskPath(depID)); os.IsNotExist(err) {
			return false
		}
		depTask, err := loadTask(depID)
		if err != nil {
			return false
		}
		if depTask.Status != "completed" {
			return false
		}
	}
	return true
}

// ClaimTask claims a task for an owner.
func ClaimTask(taskID, owner string) (string, error) {
	task, err := loadTask(taskID)
	if err != nil {
		return "", err
	}
	if task.Status != "pending" {
		return fmt.Sprintf("Task %s is %s, cannot claim", taskID, task.Status), nil
	}
	if task.Owner != nil {
		return fmt.Sprintf("Task %s already owned by %s", taskID, *task.Owner), nil
	}
	if !CanStart(taskID) {
		// Calculate missing and blocked dependencies.
		var deps []string
		var missing []string
		for _, depID := range task.BlockedBy {
			if _, err := os.Stat(taskPath(depID)); os.IsNotExist(err) {
				missing = append(missing, depID)
			} else {
				depTask, err := loadTask(depID)
				if err != nil {
					return "", err
				}
				if depTask.Status != "completed" {
					deps = append(deps, depID)
				}
			}
		}
		var parts []string
		if len(deps) > 0 {
			parts = append(parts, fmt.Sprintf("blocked by: %v", deps))
		}
		if len(missing) > 0 {
			parts = append(parts, fmt.Sprintf("missing deps: %v", missing))
		}
		return "Cannot start — " + strings.Join(parts, ", "), nil
	}
	task.Owner = &owner
	task.Status = "in_progress"
	if err := saveTask(task); err != nil {
		return "", err
	}
	fmt.Printf("  \033[36m[claim] %s → in_progress\033[0m\n", task.Subject)
	return fmt.Sprintf("Claimed %s (%s)", task.ID, task.Subject), nil
}

// CompleteTask marks a task as completed.
func CompleteTask(taskID string) (string, error) {
	task, err := loadTask(taskID)
	if err != nil {
		return "", err
	}
	if task.Status != "in_progress" {
		return fmt.Sprintf("Task %s is %s, cannot complete", taskID, task.Status), nil
	}
	task.Status = "completed"
	if err := saveTask(task); err != nil {
		return "", err
	}
	// Find unblocked tasks.
	tasks, err := ListTasks()
	if err != nil {
		return "", err
	}
	var unblocked []string
	for _, t := range tasks {
		if t.Status == "pending" && len(t.BlockedBy) > 0 && CanStart(t.ID) {
			unblocked = append(unblocked, t.Subject)
		}
	}
	fmt.Printf("  \033[32m[complete] %s ✓\033[0m\n", task.Subject)
	msg := fmt.Sprintf("Completed %s (%s)", task.ID, task.Subject)
	if len(unblocked) > 0 {
		msg += fmt.Sprintf("\nUnblocked: %s", strings.Join(unblocked, ", "))
	}
	return msg, nil
}
