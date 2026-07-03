package tools

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

func RunBash(command string) (string, error) {
	dangerous := []string{"rm -rf /", "sudo", "shutdown", "reboot", "> /dev/"}
	for _, v := range dangerous {
		if strings.Contains(command, v) {
			return "", errors.New("Dangerous command blocked")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := out.String() + stderr.String()

	if ctx.Err() == context.DeadlineExceeded {
		return "error: timeout", nil
	}
	if err != nil {
		return result, err
	}
	return result, nil
}
