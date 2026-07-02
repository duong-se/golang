package dto

// dto/project.go
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	WorkflowID  string `json:"workflow_id"`
}

// dto/workflow.go
type CreateWorkflowRequest struct {
	Name       string `json:"name"`
	StartState string `json:"start_state"`
}

type UpdateWorkflowRequest struct {
	Name       *string `json:"name"`
	StartState *string `json:"start_state"`
	Active     *bool   `json:"active"`
}

// dto/agent.go
type CreateAgentRequest struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Type     string `json:"type"`
	Active   bool   `json:"active"`
}

type AssignAgentToStateRequest struct {
	WorkflowID string `json:"workflow_id"`
	StateKey   string `json:"state_key"`
	AgentName  string `json:"agent_name"`
}
