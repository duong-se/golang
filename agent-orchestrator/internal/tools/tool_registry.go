package tools

import "encoding/json"

type ToolFunc func(args map[string]any) (string, error)

type toolRegistryImpl struct {
	tools map[string]ToolFunc
}

func NewToolRegistry() *toolRegistryImpl {
	return &toolRegistryImpl{
		tools: make(map[string]ToolFunc),
	}
}

func (r *toolRegistryImpl) Register(name string, fn ToolFunc) {
	r.tools[name] = fn
}

func (r *toolRegistryImpl) Execute(name string, args map[string]any) (string, error) {
	fn, ok := r.tools[name]
	if !ok {
		return "", nil
	}
	return fn(args)
}

func ParseArgs(raw string) (map[string]any, error) {
	var m map[string]any
	err := json.Unmarshal([]byte(raw), &m)
	return m, err
}
