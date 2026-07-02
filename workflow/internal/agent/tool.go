package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type RunResult struct {
	Success bool
	Output  string
	Error   string
}

type BasicTools struct{}

func (bt *BasicTools) RunBash(command string) *RunResult {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &RunResult{
			Success: false,
			Output:  string(output[:100]),
			Error:   fmt.Sprintf("Error: %v", err),
		}
	}
	return &RunResult{
		Success: true,
		Output:  string(output),
		Error:   "",
	}
}

func (bt *BasicTools) RunRead(path string) *RunResult {
	data, err := os.ReadFile(path)
	if err != nil {
		return &RunResult{
			Success: false,
			Output:  "",
			Error:   fmt.Sprintf("Error: %v", err),
		}
	}
	return &RunResult{
		Success: true,
		Output:  string(data),
		Error:   "",
	}
}

func (bt *BasicTools) RunWrite(path, content string) *RunResult {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return &RunResult{
			Success: false,
			Output:  "",
			Error:   fmt.Sprintf("Error: %v", err),
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return &RunResult{
			Success: false,
			Output:  "",
			Error:   fmt.Sprintf("Error: %v", err),
		}
	}
	return &RunResult{
		Success: true,
		Output:  "Wrote " + strconv.Itoa(len(content)) + " bytes",
		Error:   "",
	}
}

func NewBasicTools() *BasicTools {
	return &BasicTools{}
}
