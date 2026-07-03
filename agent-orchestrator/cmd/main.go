package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/duong-se/golang/agent-orchestrator/internal/agent"
	"github.com/duong-se/golang/agent-orchestrator/internal/providers"
	"github.com/duong-se/golang/agent-orchestrator/internal/tools"
)

func main() {
	// 1. init tool registry
	reg := tools.NewToolRegistry()

	// file tools
	reg.Register("read", func(args map[string]any) (string, error) {
		path := args["path"].(string)

		var limit *int
		if v, ok := args["limit"]; ok {
			x := int(v.(float64))
			limit = &x
		}

		return tools.RunRead(path, limit), nil
	})

	reg.Register("write", func(args map[string]any) (string, error) {
		return tools.RunWrite(
			args["path"].(string),
			args["content"].(string),
		), nil
	})

	reg.Register("edit", func(args map[string]any) (string, error) {
		return tools.RunEdit(
			args["path"].(string),
			args["old_text"].(string),
			args["new_text"].(string),
		), nil
	})

	reg.Register("glob", func(args map[string]any) (string, error) {
		return tools.RunGlob(args["pattern"].(string)), nil
	})

	// optional: bash tool
	reg.Register("bash", func(args map[string]any) (string, error) {
		return tools.RunBash(args["cmd"].(string))
	})

	// 2. OpenRouter provider (OpenAI compatible)
	p := &providers.OpenRouterProvider{
		APIKey: os.Getenv("OPENROUTER_API_KEY"),
		Model:  "openai/gpt-4o-mini",
	}

	// 3. agent init
	a := agent.New("1111", reg, p)
	a.Tools = reg

	// 4. CLI loop
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Agent ready. Type your prompt:")

	var history []agent.Message

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}
		err := scanner.Err()
		if err != nil {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "exit" {
			break
		}

		history = append(history, agent.Message{
			Role:    agent.User,
			Content: input,
		})

		result, err := a.Run(history)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}

		// print last assistant response
		for i := len(result) - 1; i >= 0; i-- {
			if result[i].Role == agent.Assistant {
				fmt.Println("\n🤖:", result[i].Content)
				break
			}
		}

		history = result
	}
}
