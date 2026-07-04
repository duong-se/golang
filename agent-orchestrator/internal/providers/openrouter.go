package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/duong-se/golang/agent-orchestrator/internal/agent"
)

type OpenRouterProvider struct {
	APIKey string
	Model  string
}

var (
	toolBlockRe = regexp.MustCompile(`(?s)<tool_call>(.*?)</tool_call>`)
	argRe       = regexp.MustCompile(`(?s)<arg_key>(.*?)</arg_key>\s*<arg_value>(.*?)</arg_value>`)
)

// ParseToolCalls parses XML-like tool calls returned by some OpenRouter models.
//
// Example:
//
//	<tool_call>shell
//	<arg_key>command</arg_key>
//	<arg_value>echo hello</arg_value>
//	<arg_key>description</arg_key>
//	<arg_value>Create file</arg_value>
//	</tool_call>
func ParseToolCalls(content string) ([]agent.ToolCall, string) {
	matches := toolBlockRe.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		return nil, strings.TrimSpace(content)
	}

	var toolCalls []agent.ToolCall

	for _, match := range matches {
		block := strings.TrimSpace(match[1])

		lines := strings.Split(block, "\n")
		if len(lines) == 0 {
			continue
		}

		toolName := strings.TrimSpace(lines[0])

		args := make(map[string]any)

		argMatches := argRe.FindAllStringSubmatch(block, -1)

		for _, arg := range argMatches {
			if len(arg) != 3 {
				continue
			}

			key := strings.TrimSpace(arg[1])
			value := strings.TrimSpace(arg[2])

			args[key] = value
		}

		argsJSON, err := json.Marshal(args)
		if err != nil {
			return toolCalls, content
		}

		toolCalls = append(toolCalls, agent.ToolCall{
			Name:      toolName,
			Arguments: string(argsJSON),
		})
	}

	// Remove tool blocks from assistant message.
	clean := toolBlockRe.ReplaceAllString(content, "")
	clean = strings.TrimSpace(clean)

	return toolCalls, clean
}

func (o *OpenRouterProvider) Name() string {
	return "openrouter"
}

func (o *OpenRouterProvider) Generate(req agent.AgentRequest) (agent.AgentResponse, error) {
	payload := map[string]any{
		"model":    o.Model,
		"messages": req.Messages,
	}

	body, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequest(
		"POST",
		"https://openrouter.ai/api/v1/chat/completions",
		bytes.NewBuffer(body),
	)

	httpReq.Header.Set("Authorization", "Bearer "+o.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("HTTP-Referer", "http://localhost") // required by OpenRouter
	httpReq.Header.Set("X-Title", "Go Agent")

	client := &http.Client{}
	res, err := client.Do(httpReq)
	if err != nil {
		return agent.AgentResponse{}, err
	}
	defer res.Body.Close()

	// var parsed struct {
	// 	Choices []struct {
	// 		Message struct {
	// 			Role     string `json:"role"`
	// 			Content  string `json:"content"`
	// 			ToolCall string `json:"tool_calls"`
	// 		} `json:"message"`
	// 	} `json:"choices"`
	// }

	var parsed = map[string]interface{}{}
	if res.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(res.Body)
		fmt.Println("ERROR:", string(bodyBytes))
		return agent.AgentResponse{}, errors.New(string(bodyBytes))
	}

	json.NewDecoder(res.Body).Decode(&parsed)
	b, _ := json.MarshalIndent(parsed, "", "  ")
	fmt.Println(string(b), "<<<<<==2222222===")
	// content := parsed.Choices[0].Message.Content

	// toolCalls, text := ParseToolCalls(content)

	// fmt.Println(text, "<<<<<=====")

	return agent.AgentResponse{
		Message: agent.Message{
			Role:    agent.Assistant,
			Content: "",
			// Content:   text,
			// ToolCalls: toolCalls,
		},
	}, nil
}
