package domain

// domain/agent.go
type Agent struct {
	ID       string
	Name     string
	Endpoint string
	Type     string
	Active   bool
}

// domain/workflow.go
type Workflow struct {
	ID         string
	Name       string
	Version    int
	StartState string
	Active     bool
}

// domain/project.go
type Project struct {
	ID          string
	Name        string
	Description string
	WorkflowID  string
	Status      string
}
