package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

func RunWrite(path string, content string) string {
	filePath, err := SafePath(path)
	if err != nil {
		return "Error: " + err.Error()
	}

	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return "Error: " + err.Error()
	}

	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "Error: " + err.Error()
	}

	return fmt.Sprintf("Wrote %d bytes to %s", len(content), path)
}
