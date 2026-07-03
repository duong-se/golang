package agent

import "github.com/duong-se/golang/agent-orchestrator/internal/tools"

type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
	Tool      Role = "tool"
)

type Message struct {
	Role      Role       `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type AgentRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type AgentResponseContent struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Input string `json:"input"`
}

type AgentResponse struct {
	Message    Message              `json:"message"`
	StopReason string               `json:"stop_reason"`
	Content    AgentResponseContent `json:"content"`
	// Raw        any     `json:"raw,omitempty"`
}

type Provider interface {
	Name() string
	Generate(req AgentRequest) (AgentResponse, error)
}

type ToolCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type agentImpl struct {
	UUID     string
	Name     string
	Tools    ToolRegistry
	Provider Provider
}

type ToolRegistry interface {
	Execute(name string, args map[string]any) (string, error)
}

func New(uuid string, tools ToolRegistry, provider Provider) *agentImpl {
	return &agentImpl{
		UUID:     uuid,
		Tools:    tools,
		Provider: provider,
	}
}

func (a *agentImpl) isDone(resp AgentResponse) bool {
	return len(resp.Message.ToolCalls) == 0
}

func (a *agentImpl) handleTools(resp AgentResponse) (*Message, error) {
	if len(resp.Message.ToolCalls) == 0 {
		return nil, nil
	}
	results := ""
	for _, call := range resp.Message.ToolCalls {
		args, err := tools.ParseArgs(call.Arguments)
		if err != nil {
			results += "[tool parse error]\n"
			continue
		}
		out, err := a.Tools.Execute(call.Name, args)
		if err != nil {
			results += call.Name + ": error\n"
			continue
		}
		results += call.Name + ": " + out + "\n"
	}

	return &Message{
		Role:    Tool,
		Content: results,
	}, nil
}

func (a *agentImpl) Run(initial []Message) ([]Message, error) {
	messages := append([]Message{}, initial...)
	for {
		resp, err := a.Provider.Generate(AgentRequest{
			Messages: messages,
		})
		if err != nil {
			return nil, err
		}
		messages = append(messages, Message{
			Role:    resp.Message.Role,
			Content: resp.Message.Content,
		})
		if resp.StopReason != "tool_use" {
			return nil, nil
		}
		toolMsg, err := a.handleTools(resp)
		if err != nil {
			return nil, err
		}
		if toolMsg != nil {
			messages = append(messages, *toolMsg)
			continue
		}
		if a.isDone(resp) {
			return messages, nil
		}
	}
}
