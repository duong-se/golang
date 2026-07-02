package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	MicroCompactThreshold   = 120
	KeepRecentToolResults   = 3
	TranscriptDirName       = ".transcripts"
	ToolResultsDirName      = ".task_outputs/tool-results"
	KeepRecentToolThreshold = 3
	MaxCompactMessages      = 50
	ToolResultBudgetBytes   = 200_000
)

// SystemConfig holds configuration for compaction
type SystemConfig struct {
	WorkDir string
}

// EstimateSize estimates the size of a message list in bytes
func EstimateSize(messages []Message) int {
	data, _ := json.Marshal(messages)
	return len(data)
}

// BlockType returns the type of a block (tool_use or text)
func BlockType(block map[string]interface{}) string {
	if t, ok := block["type"].(string); ok {
		return t
	}
	return ""
}

// MessageHasToolUse checks if a message has tool use
func MessageHasToolUse(msg Message) bool {
	if msg.Role != "assistant" {
		return false
	}
	if contentList, ok := msg.Content.([]interface{}); ok {
		for _, item := range contentList {
			if block, ok := item.(map[string]interface{}); ok {
				if BlockType(block) == "tool_use" {
					return true
				}
			}
		}
	}
	return false
}

// IsToolResultMessage checks if a message is a tool result
func IsToolResultMessage(msg Message) bool {
	if msg.Role != "user" {
		return false
	}
	if contentList, ok := msg.Content.([]interface{}); ok {
		for _, item := range contentList {
			if block, ok := item.(map[string]interface{}); ok {
				if BlockType(block) == "tool_result" {
					return true
				}
			}
		}
	}
	return false
}

// CollectToolResults collects all tool result blocks from messages
func CollectToolResults(messages []Message) [][]interface{} {
	var results [][]interface{}
	for msgIdx, msg := range messages {
		if msg.Role != "user" {
			continue
		}
		if contentList, ok := msg.Content.([]interface{}); ok {
			for blockIdx, blockItem := range contentList {
				if block, ok := blockItem.(map[string]interface{}); ok {
					if BlockType(block) == "tool_result" {
						results = append(results, []interface{}{msgIdx, blockIdx, block})
					}
				}
			}
		}
	}
	return results
}

// PersistLargeOutput returns a persisted output if the output is too large
func PersistLargeOutput(toolUseID string, output string) string {
	if len(output) <= PERSIST_THRESHOLD {
		return output
	}
	transcriptDir := filepath.Join(WorkDir, TranscriptDirName)
	_ = os.MkdirAll(transcriptDir, 0755)
	path := filepath.Join(transcriptDir, toolUseID+".txt")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.WriteFile(path, []byte(output), 0644)
	}
	return fmt.Sprintf("<persisted-output>\nFull output: %s\nPreview:\n%s\n</persisted-output>", path, output[:2000])
}

// ToolResultBudget ensures tool results don't exceed a byte limit by persisting large ones
func ToolResultBudget(messages []Message, maxBytes int) []Message {
	if len(messages) == 0 {
		return messages
	}
	last := messages[len(messages)-1]
	if last.Role != "user" {
		return messages
	}
	if contentList, ok := last.Content.([]interface{}); !ok || len(contentList) == 0 {
		return messages
	}

	var blocks [][]interface{}
	for i, b := range contentList {
		if block, ok := b.(map[string]interface{}); ok {
			if BlockType(block) == "tool_result" {
				blocks = append(blocks, []interface{}{i, b})
			}
		}
	}

	total := 0
	for _, pair := range blocks {
		if block, ok := pair[1].(map[string]interface{}); ok {
			if content, ok := block["content"].(string); ok {
				total += len(content)
			}
		}
	}

	if total <= maxBytes {
		return messages
	}

	sort.Slice(blocks, func(i, j int) bool {
		lenI := 0
		if blockI, ok := blocks[i][1].(map[string]interface{}); ok {
			if contentI, ok := blockI["content"].(string); ok {
				lenI = len(contentI)
			}
		}
		lenJ := 0
		if blockJ, ok := blocks[j][1].(map[string]interface{}); ok {
			if contentJ, ok := blockJ["content"].(string); ok {
				lenJ = len(contentJ)
			}
		}
		return lenI > lenJ
	})

	for _, pair := range blocks {
		if total <= maxBytes {
			break
		}
		if block, ok := pair[1].(map[string]interface{}); ok {
			if content, ok := block["content"].(string); ok {
				persisted := PersistLargeOutput(
					fmt.Sprintf("%v", block["tool_use_id"]), content)
				block["content"] = persisted
				if newContent, ok := block["content"].(string); ok {
					total = total - len(content) + len(newContent)
				}
			}
		}
	}
	return messages
}

// SnipCompact removes middle section of messages if there are too many
func SnipCompact(messages []Message, maxMessages int) []Message {
	if len(messages) <= maxMessages {
		return messages
	}

	headEnd, tailStart := 3, len(messages)-(maxMessages-3)
	if headEnd > 0 && MessageHasToolUse(messages[headEnd-1]) {
		for headEnd < len(messages) && IsToolResultMessage(messages[headEnd]) {
			headEnd++
		}
	}
	if tailStart > 0 && tailStart < len(messages) && IsToolResultMessage(messages[tailStart]) && MessageHasToolUse(messages[tailStart-1]) {
		tailStart--
	}

	if headEnd >= tailStart {
		return messages
	}

	snipped := tailStart - headEnd
	var result []Message
	result = append(result, messages[:headEnd]...)
	result = append(result, Message{
		Role:    "user",
		Content: fmt.Sprintf("[snipped %d messages]", snipped),
	})
	result = append(result, messages[tailStart:]...)
	return result
}

// MicroCompact compacts old tool results
func MicroCompact(messages []Message) []Message {
	toolResults := CollectToolResults(messages)
	if len(toolResults) <= KeepRecentToolResults {
		return messages
	}

	for _, pair := range toolResults[:len(toolResults)-KeepRecentToolResults] {
		if msgIdx, ok := pair[0].(int); ok {
			if blockIdx, ok := pair[1].(int); ok {
				if messages[msgIdx].Role == "user" {
					if contentList, ok := messages[msgIdx].Content.([]interface{}); ok && len(contentList) > blockIdx {
						if block, ok := contentList[blockIdx].(map[string]interface{}); ok {
							if content, ok := block["content"].(string); ok && len(content) > MicroCompactThreshold {
								block["content"] = "[Earlier tool result compacted. Re-run if needed.]"
								contentList[blockIdx] = block
								messages[msgIdx].Content = contentList
							}
						}
					}
				}
			}
		}
	}
	return messages
}

// WriteTranscript writes messages to a transcript file
func WriteTranscript(messages []Message) string {
	_ = os.MkdirAll(filepath.Join(WorkDir, TranscriptDirName), 0755)
	path := filepath.Join(WorkDir, TranscriptDirName, fmt.Sprintf("transcript_%d.jsonl", time.Now().Unix()))
	data, _ := json.Marshal(messages)
	_ = os.WriteFile(path, data, 0644)
	return path
}

// SummarizeHistory creates a summary of conversation history
func SummarizeHistory(messages []Message) string {
	maxLength := 80000
	if len(messages) == 0 {
		return "(empty summary)"
	}

	var conversation []interface{}
	for _, msg := range messages {
		if len(conversation) > maxLength/1024 {
			break
		}
		conversation = append(conversation, msg)
	}

	data, _ := json.Marshal(conversation)
	if len(data) > maxLength {
		data = data[:maxLength]
	}

	prompt := fmt.Sprintf("Summarize this coding-agent conversation so work can continue. Preserve current goal, key findings, changed files, remaining work, and user constraints.\n\n%s", string(data))

	// In a real implementation, this would call the LLM
	// For now, return a placeholder
	return "Summary of conversation (would be generated by LLM in production)"
}

// CompactHistory creates a compacted version of the conversation
func CompactHistory(messages []Message) []Message {
	transcript := WriteTranscript(messages)
	fmt.Printf("  \033[36m[compact] transcript saved: %s\033[0m\n", transcript)

	summary := SummarizeHistory(messages)
	return []Message{
		{
			Role:    "user",
			Content: fmt.Sprintf("[Compacted]\n\n%s", summary),
		},
	}
}

// ReactiveCompact creates a reactive compacted version of the conversation
func ReactiveCompact(messages []Message) []Message {
	transcript := WriteTranscript(messages)
	fmt.Printf("  \033[31m[reactive compact] transcript saved: %s\033[0m\n", transcript)

	tailStart := max(0, len(messages)-5)
	if tailStart > 0 && tailStart < len(messages) && IsToolResultMessage(messages[tailStart]) && MessageHasToolUse(messages[tailStart-1]) {
		tailStart--
	}

	summary := "Earlier conversation was trimmed after a prompt-too-long error."
	if tailStart > 0 {
		summary = SummarizeHistory(messages[:tailStart])
	}

	var result []Message
	result = append(result, Message{
		Role:    "user",
		Content: fmt.Sprintf("[Reactive compact]\n\n%s", summary),
	})
	result = append(result, messages[tailStart:]...)
	return result
}
