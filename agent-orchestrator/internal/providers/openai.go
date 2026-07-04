package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/duong-se/golang/agent-orchestrator/internal/agent"
)

type OpenAIProvider struct {
	APIKey string
	Model  string
}

/*
========================
OpenAI Request Schema
========================
*/
type openAIRequest struct {
	Model    string        `json:"model"`
	Messages []interface{} `json:"messages"`
}

/*
========================
OpenAI Response Schema
========================
*/
type openAIResponse struct {
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

/*
========================
Provider Name
========================
*/
func (o *OpenAIProvider) Name() string {
	return "openai"
}

/*
========================
Generate
========================
*/
func (o *OpenAIProvider) Generate(req agent.AgentRequest) (agent.AgentResponse, error) {
	payload := map[string]any{
		"model":    o.Model,
		"messages": req.Messages,
	}

	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return agent.AgentResponse{}, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+o.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

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

	var parsed openAIResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return agent.AgentResponse{}, err
	}

	if len(parsed.Choices) == 0 {
		return agent.AgentResponse{}, errors.New("no choices returned")
	}

	msg := parsed.Choices[0].Message

	toolCalls := make([]agent.ToolCall, 0, len(msg.ToolCalls))

	for _, tc := range msg.ToolCalls {
		toolCalls = append(toolCalls, agent.ToolCall{
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return agent.AgentResponse{
		Message: agent.Message{
			Role:      agent.Role(msg.Role),
			Content:   msg.Content,
			ToolCalls: toolCalls,
		},
	}, nil
}
