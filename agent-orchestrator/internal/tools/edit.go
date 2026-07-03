package tools

import (
	"os"
	"strings"
)

func RunEdit(path string, oldText string, newText string) string {
	filePath, err := SafePath(path)
	if err != nil {
		return "Error: " + err.Error()
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "Error: " + err.Error()
	}

	text := string(data)

	if !strings.Contains(text, oldText) {
		return "Error: text not found in " + path
	}

	text = strings.Replace(text, oldText, newText, 1)

	err = os.WriteFile(filePath, []byte(text), 0644)
	if err != nil {
		return "Error: " + err.Error()
	}

	return "Edited " + path
}
