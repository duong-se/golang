package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/duong-se/golang/agent-orchestrator/internal/agent"
)

type GeminiProvider struct {
	APIKey string
	Model  string
}

/*
========================
Gemini Request Schema
========================
*/
type geminiRequest struct {
	Contents []struct {
		Role  string `json:"role"`
		Parts []any  `json:"parts"`
	} `json:"contents"`
}

/*
========================
Gemini Response Schema
========================
*/
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Role  string `json:"role"`
			Parts []struct {
				Text string `json:"text,omitempty"`

				FunctionCall *struct {
					Name string         `json:"name"`
					Args map[string]any `json:"args"`
				} `json:"functionCall,omitempty"`
			} `json:"parts"`
		} `json:"content"`

		FinishReason string `json:"finishReason"`
	} `json:"candidates"`

	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

/*
========================
Provider Name
========================
*/
func (g *GeminiProvider) Name() string {
	return "gemini"
}

/*
========================
Generate
========================
*/
func (g *GeminiProvider) Generate(req agent.AgentRequest) (agent.AgentResponse, error) {

	payload := geminiRequest{
		Contents: []struct {
			Role  string `json:"role"`
			Parts []any  `json:"parts"`
		}{},
	}

	// map messages → Gemini format (simple passthrough)
	for _, m := range req.Messages {
		payload.Contents = append(payload.Contents, struct {
			Role  string `json:"role"`
			Parts []any  `json:"parts"`
		}{
			Role:  string(m.Role),
			Parts: []any{map[string]any{"text": m.Content}},
		})
	}

	body, _ := json.Marshal(payload)

	url := "https://generativelanguage.googleapis.com/v1beta/models/" +
		g.Model + ":generateContent?key=" + g.APIKey

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return agent.AgentResponse{}, err
	}

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

	var parsed geminiResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return agent.AgentResponse{}, err
	}

	if len(parsed.Candidates) == 0 {
		return agent.AgentResponse{}, errors.New("no candidates returned")
	}

	var (
		text      string
		toolCalls []agent.ToolCall
	)

	for _, part := range parsed.Candidates[0].Content.Parts {

		// text part
		if part.Text != "" {
			text += part.Text
		}

		// function call part
		if part.FunctionCall != nil {

			b, _ := json.Marshal(part.FunctionCall.Args)

			toolCalls = append(toolCalls, agent.ToolCall{
				Name:      part.FunctionCall.Name,
				Arguments: string(b),
			})
		}
	}

	return agent.AgentResponse{
		Message: agent.Message{
			Role:      agent.Role(parsed.Candidates[0].Content.Role),
			Content:   text,
			ToolCalls: toolCalls,
		},
	}, nil
}
