package agent

type Hook func(args ...interface{}) (interface{}, bool)

var hooks = struct {
	UserPromptSubmit []Hook
	PreToolUse       []Hook
	PostToolUse      []Hook
	Stop             []Hook
}{}

func RegisterHook(event string, hook Hook) {
	switch event {
	case "UserPromptSubmit":
		hooks.UserPromptSubmit = append(hooks.UserPromptSubmit, hook)
	case "PreToolUse":
		hooks.PreToolUse = append(hooks.PreToolUse, hook)
	case "PostToolUse":
		hooks.PostToolUse = append(hooks.PostToolUse, hook)
	case "Stop":
		hooks.Stop = append(hooks.Stop, hook)
	}
}

func TriggerHooks(event string, args ...interface{}) (interface{}, bool) {
	var targetHooks []Hook
	switch event {
	case "UserPromptSubmit":
		targetHooks = hooks.UserPromptSubmit
	case "PreToolUse":
		targetHooks = hooks.PreToolUse
	case "PostToolUse":
		targetHooks = hooks.PostToolUse
	case "Stop":
		targetHooks = hooks.Stop
	}
	for _, h := range targetHooks {
		if res, ok := h(args...); ok {
			return res, true
		}
	}
	return nil, false
}
