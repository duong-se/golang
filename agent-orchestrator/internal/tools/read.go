package tools

import (
	"fmt"
	"os"
	"strings"
)

func RunRead(path string, limit *int) string {
	filePath, err := SafePath(path)
	if err != nil {
		return "Error: " + err.Error()
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "Error: " + err.Error()
	}

	lines := strings.Split(string(data), "\n")

	if limit != nil && *limit < len(lines) {
		lines = append(lines[:*limit],
			"... ("+fmt.Sprintf("%d", len(lines)-*limit)+" more lines)")
	}

	return strings.Join(lines, "\n")
}
