package agent

import (
	"fmt"
	"strings"
	"time"
)

const (
	DEFAULT_MAX_TOKENS       = 8000
	ESCALATED_MAX_TOKENS     = 16000
	MAX_RETRIES              = 3
	MAX_CONSECUTIVE_529      = 2
	MAX_RECOVERY_RETRIES     = 2
	BASE_DELAY_MS            = 500
	CONTEXT_LIMIT            = 50000
	KEEP_RECENT_TOOL_RESULTS = 3
	PERSIST_THRESHOLD        = 30000
	CONTINUATION_PROMPT      = "Continue from the previous response. Do not repeat completed work."
	IDLE_POLL_INTERVAL       = 5
	IDLE_TIMEOUT             = 60
)

var (
	MODEL          string
	PRIMARY_MODEL  string
	FALLBACK_MODEL string
	CurrentTodos   []map[string]interface{}
)

var PromptSections = map[string]string{
	"identity": "You are a coding agent. Act, don't explain.",
	"tools": "Available tools: bash, read_file, write_file, edit_file, glob, " +
		"todo_write, task, load_skill, compact, " +
		"create_task, list_tasks, get_task, claim_task, complete_task, " +
		"schedule_cron, list_crons, cancel_cron, " +
		"spawn_teammate, send_message, check_inbox, " +
		"request_shutdown, request_plan, review_plan, " +
		"create_worktree, remove_worktree, keep_worktree, " +
		"connect_mcp. MCP tools are prefixed mcp__{server}__{tool}.",
	"workspace": "Working directory: " + WorkDir,
	"memory":    "Relevant memories are injected below when available.",
}

func AssembleSystemPrompt(context map[string]interface{}) string {
	sections := []string{
		PromptSections["identity"],
		PromptSections["tools"],
		PromptSections["workspace"],
	}
	sections = append(sections, fmt.Sprintf("Current time: %s", time.Now().Format(time.RFC3339)))
	sections = append(sections, "Skills catalog:\n"+ListSkills()+"\nUse load_skill(name) when a skill is relevant.")

	if memories, ok := context["memories"].(string); ok && memories != "" {
		sections = append(sections, fmt.Sprintf("Relevant memories:\n%s", memories))
	}

	return strings.Join(sections, "\n\n")
}
