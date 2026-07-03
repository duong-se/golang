package main

import (
	"github.com/duong-se/golang/agent-orchestrator/internal/tools"
)

func main() {
	agentTools := tools.NewToolRegistry()
	agentTools.Register("bash", func(args map[string]any) (string, error) {
		cmdStr, ok := args["cmd"].(string)
		if !ok {
			return "", nil
		}
		return tools.RunBash(cmdStr)
	})

	agentTools.Register("read", func(args map[string]any) (string, error) {
		path := args["path"].(string)

		var limit *int
		if v, ok := args["limit"]; ok {
			x := int(v.(float64))
			limit = &x
		}

		return tools.RunRead(path, limit), nil
	})

	agentTools.Register("write", func(args map[string]any) (string, error) {
		return tools.RunWrite(
			args["path"].(string),
			args["content"].(string),
		), nil
	})

	agentTools.Register("edit", func(args map[string]any) (string, error) {
		return tools.RunEdit(
			args["path"].(string),
			args["old_text"].(string),
			args["new_text"].(string),
		), nil
	})

	agentTools.Register("glob", func(args map[string]any) (string, error) {
		return tools.RunGlob(args["pattern"].(string)), nil
	})
}
