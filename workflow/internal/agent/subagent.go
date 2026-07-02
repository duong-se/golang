package agent

import (
	"context"
	"strings"
)

const (
	MaxSubagentIterations = 30
)

type Subagent struct {
	Description string
	Context     context.Context
	Messages    []Message
	Completed   chan string
}

type SubagentConfig struct {
	SystemPrompt string
	Tools        []ToolDefinition
	MaxTokens    int
}

type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

func NewSubagent(description string) *Subagent {
	return &Subagent{
		Description: description,
		Messages:    []Message{{Role: "user", Content: description}},
		Completed:   make(chan string, 1),
	}
}

func (s *Subagent) Run(config SubagentConfig) string {
	for i := 0; i < MaxSubagentIterations; i++ {
		response := s.callModel(config)
		if response == nil {
			break
		}

		s.Messages = append(s.Messages, Message{
			Role:    "assistant",
			Content: response.Content,
		})

		if !s.hasToolUse(response.Content) {
			break
		}

		results := s.executeTools(response, config)
		s.Messages = append(s.Messages, Message{
			Role:    "user",
			Content: results,
		})
	}

	// Return final summary
	for i := len(s.Messages) - 1; i >= 0; i-- {
		if s.Messages[i].Role == "assistant" {
			text := s.extractText(s.Messages[i].Content)
			if text != "" {
				return text
			}
		}
	}

	return "Subagent finished without a text summary."
}

func (s *Subagent) callModel(config SubagentConfig) *ModelResponse {
	// Placeholder - in real implementation this would call the LLM
	return &ModelResponse{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": "Subagent completed task",
			},
		},
	}
}

func (s *Subagent) hasToolUse(content interface{}) bool {
	if contentList, ok := content.([]interface{}); ok {
		for _, item := range contentList {
			if block, ok := item.(map[string]interface{}); ok {
				if block["type"] == "tool_use" {
					return true
				}
			}
		}
	}
	return false
}

func (s *Subagent) executeTools(response *ModelResponse, config SubagentConfig) []interface{} {
	var results []interface{}

	if contentList, ok := response.Content.([]interface{}); ok {
		for _, item := range contentList {
			if block, ok := item.(map[string]interface{}); ok {
				if block["type"] == "tool_use" {
					toolName := block["name"].(string)
					toolInput := block["input"].(map[string]interface{})
					toolUseID := block["id"].(string)

					result := s.executeTool(toolName, toolInput, config)
					results = append(results, map[string]interface{}{
						"type":        "tool_result",
						"tool_use_id": toolUseID,
						"content":     result,
					})
				}
			}
		}
	}

	return results
}

func (s *Subagent) executeTool(toolName string, toolInput map[string]interface{}, config SubagentConfig) string {
	// In real implementation, this would call actual tool handlers
	switch toolName {
	case "bash":
		return "Command executed"
	case "read_file":
		return "File read"
	case "write_file":
		return "File written"
	case "edit_file":
		return "File edited"
	case "glob":
		return "Glob executed"
	default:
		return "Unknown tool"
	}
}

func (s *Subagent) extractText(content interface{}) string {
	if contentList, ok := content.([]interface{}); ok {
		var texts []string
		for _, item := range contentList {
			if block, ok := item.(map[string]interface{}); ok {
				if block["type"] == "text" {
					if text, ok := block["text"].(string); ok {
						texts = append(texts, text)
					}
				}
			}
		}
		return strings.Join(texts, "\n")
	}
	return ""
}

type ModelResponse struct {
	Content interface{} `json:"content"`
}
