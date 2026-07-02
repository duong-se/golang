package agent

import (
	"os"
	"path/filepath"
)

type Agent struct {
	Config   Config
	WorkDir  string
	Messages []Message
	Running  bool
}

func New() *Agent {
	wd, _ := os.Getwd()
	return &Agent{
		Config:  LoadConfig(),
		WorkDir: filepath.Clean(wd),
		Running: true,
	}
}
