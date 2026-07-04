package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/duong-se/golang/agent-orchestrator/internal/agent"
)

type ClaudeProvider struct {
	APIKey string
	Model  string
}

/*
========================
Claude Request Schema
========================
*/
type claudeRequest struct {
	Model     string        `json:"model"`
	Messages  []interface{} `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
}

/*
========================
Claude Response Schema
========================
*/
type claudeResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	Content []struct {
		Type string `json:"type"`

		Text string `json:"text,omitempty"`

		// tool_use
		ID    string         `json:"id,omitempty"`
		Name  string         `json:"name,omitempty"`
		Input map[string]any `json:"input,omitempty"`
	} `json:"content"`

	Role       string `json:"role"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`

	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

/*
========================
Provider Name
========================
*/
func (c *ClaudeProvider) Name() string {
	return "claude"
}

/*
========================
Generate
========================
*/
func (c *ClaudeProvider) Generate(req agent.AgentRequest) (agent.AgentResponse, error) {
	payload := map[string]any{
		"model":      c.Model,
		"messages":   req.Messages,
		"max_tokens": 4096,
	}

	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequest(
		"POST",
		"https://api.anthropic.com/v1/messages",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return agent.AgentResponse{}, err
	}

	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	client := &http.Client{}
	res, err := client.Do(httpReq)
	if err != nil {
		return agent.AgentResponse{}, err
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return agent.AgentResponse{}, errors.New(string(raw))
	}

	var parsed claudeResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return agent.AgentResponse{}, err
	}

	var (
		text      string
		toolCalls []agent.ToolCall
	)

	for _, block := range parsed.Content {

		switch block.Type {

		case "text":
			text += block.Text

		case "tool_use":
			b, _ := json.Marshal(block.Input)

			toolCalls = append(toolCalls, agent.ToolCall{
				Name:      block.Name,
				Arguments: string(b),
			})
		}
	}

	return agent.AgentResponse{
		Message: agent.Message{
			Role:      agent.Role(parsed.Role),
			Content:   text,
			ToolCalls: toolCalls,
		},
	}, nil
}
