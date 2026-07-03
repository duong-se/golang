package agent

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

type AgentResponse struct {
	Message Message `json:"message"`
	Raw     any     `json:"raw,omitempty"`
}

type Provider interface {
	Name() string
	Generate(req AgentRequest) (AgentResponse, error)
}

type ToolCall struct {
	Command string `json:"command"`
}

type agentImpl struct {
	UUID     string
	Name     string
	Tools    ToolRegistry
	Provider Provider
}

type ToolRegistry interface {
	Execute(command string) (string, error)
}

func New(uuid string, model string, tools ToolRegistry, provider Provider) *agentImpl {
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
		out, err := a.Tools.Execute(call.Command)
		if err != nil {
			results += call.Command + ": error\n"
			continue
		}
		results += call.Command + ": " + out + "\n"
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
		messages = append(messages, resp.Message)
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
