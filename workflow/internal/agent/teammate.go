package agent

import (
	"context"
	"fmt"
	"os"
)

type Teammate struct {
	Name      string
	Role      string
	Context   context.Context
	Worktree  string
	completed chan struct{}
}

type TeammateFunc func(context.Context) error

func spawnTeammate(name, role string, prompt string) *Teammate {
	return &Teammate{
		Name:      name,
		Role:      role,
		completed: make(chan struct{}),
	}
}

func (t *Teammate) runInner(ctx context.Context, fn TeammateFunc) {
	err := fn(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Teammate %s failed: %v\n", t.Name, err)
	}
	t.completed <- struct{}{}
}

func (t *Teammate) Run(fn TeammateFunc) {
	go func() {
		t.runInner(t.Context, fn)
	}()
}

func (t *Teammate) WaitForCompletion() {
	<-t.completed
}
