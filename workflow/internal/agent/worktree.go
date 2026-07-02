package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// WorktreeDir is the directory where worktrees are stored.
var WorktreeDir string

// validWorktreeNameRegex is the regex for valid worktree names.
var validWorktreeNameRegex = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	WorktreeDir = filepath.Join(wd, ".worktrees")
	if err := os.MkdirAll(WorktreeDir, 0755); err != nil {
		panic(err)
	}
}

// ValidateWorktreeName validates a worktree name.
func ValidateWorktreeName(name string) string {
	if name == "" {
		return "Worktree name cannot be empty"
	}
	if name == "." || name == ".." {
		return fmt.Sprintf("'%s' is not a valid worktree name", name)
	}
	if !validWorktreeNameRegex.MatchString(name) {
		return fmt.Sprintf("Invalid worktree name '%s': only letters, digits, dots, underscores, dashes (1-64 chars)", name)
	}
	return ""
}

// runGit runs a git command and returns (success, output).
func runGit(args ...string) (bool, string) {
	// We'll use the working directory as the base.
	// In the Python version, they use WORKDIR which is the current working directory.
	// We'll do the same.
	wd, err := os.Getwd()
	if err != nil {
		return false, "Error: cannot get working directory"
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(output))
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return false, outStr[:5000]
		}
		return false, "Error: git timeout"
	}
	if outStr == "" {
		outStr = "(no output)"
	}
	return true, outStr[:5000]
}

// logEvent logs an event to the worktree events file.
func logEvent(eventType, worktreeName, taskID string) {
	event := map[string]interface{}{
		"type":     eventType,
		"worktree": worktreeName,
		"task_id":  taskID,
		"ts":       time.Now().Unix(),
	}
	data, _ := json.Marshal(event)
	eventsFile := filepath.Join(WorktreeDir, "events.jsonl")
	f, err := os.OpenFile(eventsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(string(data) + "\n")
}

// CreateWorktree creates a new worktree.
func CreateWorktree(name, taskID string) string {
	if err := ValidateWorktreeName(name); err != "" {
		return "Error: " + err
	}
	if taskID != "" {
		if _, err := loadTask(taskID); err != nil {
			return "Error: task " + taskID + " not found"
		}
	}
	path := filepath.Join(WorktreeDir, name)
	if _, err := os.Stat(path); err == nil {
		return fmt.Sprintf("Worktree '%s' already exists at %s", name, path)
	}
	ok, result := runGit("worktree", "add", path, "-b", fmt.Sprintf("wt/%s", name), "HEAD")
	if !ok {
		return "Git error: " + result
	}
	if taskID != "" {
		if err := bindTaskToWorktree(taskID, name); err != nil {
			return err.Error()
		}
	}
	logEvent("create", name, taskID)
	fmt.Printf("  \033[33m[worktree] created: %s at %s\033[0m\n", name, path)
	return fmt.Sprintf("Worktree '%s' created at %s", name, path)
}

// bindTaskToWorktree binds a task to a worktree.
func bindTaskToWorktree(taskID, worktreeName string) error {
	task, err := loadTask(taskID)
	if err != nil {
		return err
	}
	task.Worktree = &worktreeName
	return saveTask(task)
}

// countWorktreeChanges counts the number of changed files and commits in a worktree.
func countWorktreeChanges(path string) (int, int) {
	wd, err := os.Getwd()
	if err != nil {
		return -1, -1
	}
	// git status --porcelain
	cmd1 := exec.Command("git", "status", "--porcelain")
	cmd1.Dir = path
	output1, err := cmd1.Output()
	if err != nil {
		return -1, -1
	}
	files := 0
	for _, line := range strings.Split(strings.TrimSpace(string(output1)), "\n") {
		if strings.TrimSpace(line) != "" {
			files++
		}
	}
	// git log @{push}..HEAD --oneline
	cmd2 := exec.Command("git", "log", "@{push}..HEAD", "--oneline")
	cmd2.Dir = path
	output2, err := cmd2.Output()
	if err != nil {
		return -1, -1
	}
	commits := 0
	for _, line := range strings.Split(strings.TrimSpace(string(output2)), "\n") {
		if strings.TrimSpace(line) != "" {
			commits++
		}
	}
	return files, commits
}

// RemoveWorktree removes a worktree.
func RemoveWorktree(name string, discardChanges bool) string {
	if err := ValidateWorktreeName(name); err != "" {
		return err
	}
	path := filepath.Join(WorktreeDir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Sprintf("Worktree '%s' not found", name)
	}
	if !discardChanges {
		files, commits := countWorktreeChanges(path)
		if files < 0 {
			return "Cannot verify status. Use discard_changes=true to force."
		}
		if files > 0 || commits > 0 {
			return fmt.Sprintf("Worktree '%s' has %d file(s), %d commit(s). Use discard_changes=true or keep_worktree.", name, files, commits)
		}
	}
	ok1, _ := runGit("worktree", "remove", path, "--force")
	if !ok1 {
		return fmt.Sprintf("Failed to remove worktree '%s'", name)
	}
	runGit("branch", "-D", fmt.Sprintf("wt/%s", name))
	logEvent("remove", name, "")
	fmt.Printf("  \033[33m[worktree] removed: %s\033[0m\n", name)
	return fmt.Sprintf("Worktree '%s' removed", name)
}

// KeepWorktree marks a worktree as kept for review.
func KeepWorktree(name string) string {
	if err := ValidateWorktreeName(name); err != "" {
		return err
	}
	logEvent("keep", name, "")
	return fmt.Sprintf("Worktree '%s' kept for review (branch: wt/%s)", name, name)
}
