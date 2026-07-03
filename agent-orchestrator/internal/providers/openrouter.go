package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/duong-se/golang/agent-orchestrator/internal/agent"
)

type OpenRouterProvider struct {
	APIKey string
	Model  string
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

	var parsed map[string]interface{}

	if res.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(res.Body)
		fmt.Println("ERROR:", string(bodyBytes))
		return agent.AgentResponse{}, errors.New(string(bodyBytes))
	}

	json.NewDecoder(res.Body).Decode(&parsed)
	fmt.Println(parsed, "<<<<<=====")
	// if len(parsed.Choices) < 1 {
	// 	return agent.AgentResponse{
	// 		Message: agent.Message{
	// 			Role:    agent.Assistant,
	// 			Content: "",
	// 		},
	// 	}, nil
	// }
	return agent.AgentResponse{
		Message: agent.Message{
			Role:    agent.Assistant,
			Content: "",
		},
	}, nil
}
