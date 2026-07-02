package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var WorkDir string
var SkillsDir string

var SkillRegistry = make(map[string]map[string]interface{})

var transcriptDir string
var toolResultsDir string
var mailboxDir string

func init() {
	WorkDir, _ = os.Getwd()
	SkillsDir = filepath.Join(WorkDir, "skills")
	transcriptDir = filepath.Join(WorkDir, ".transcripts")
	toolResultsDir = filepath.Join(WorkDir, ".task_outputs", "tool-results")
	mailboxDir = filepath.Join(WorkDir, ".mailboxes")

	os.MkdirAll(filepath.Join(WorkDir, ".tasks"), 0755)
	os.MkdirAll(filepath.Join(WorkDir, ".worktrees"), 0755)
	os.MkdirAll(mailboxDir, 0755)
}

// ParseFrontmatter parses markdown frontmatter
func ParseFrontmatter(text string) (map[string]interface{}, string) {
	if !strings.HasPrefix(text, "---") {
		return map[string]interface{}{}, text
	}
	parts := strings.SplitN(text, "---", 3)
	if len(parts) < 3 {
		return map[string]interface{}{}, text
	}
	meta := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		meta = map[string]interface{}{}
	}
	return meta, strings.TrimSpace(parts[2])
}

// ScanSkills scans the skills directory and registers skill manifests
func ScanSkills() {
	SkillRegistry = make(map[string]map[string]interface{})
	entries, err := ioutil.ReadDir(SkillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		return
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)

	for _, dir := range dirs {
		manifest := filepath.Join(SkillsDir, dir, "SKILL.md")
		if !fileExists(manifest) {
			continue
		}
		raw, err := ioutil.ReadFile(manifest)
		if err != nil {
			continue
		}
		meta, _ := ParseFrontmatter(string(raw))
		name := meta["name"].(string)
		if name == "" {
			name = dir
		}
		desc := meta["description"].(string)
		if desc == "" {
			desc = "(no description)"
		}
		SkillRegistry[name] = map[string]interface{}{
			"name":        name,
			"description": desc,
			"content":     string(raw),
		}
	}
}

// ListSkills returns formatted list of skill names and descriptions
func ListSkills() string {
	if len(SkillRegistry) == 0 {
		return "(no skills found)"
	}
	var parts []string
	for _, skill := range SkillRegistry {
		parts = append(parts, fmt.Sprintf("- %s: %s", skill["name"], skill["description"]))
	}
	return strings.Join(parts, "\n")
}

// LoadSkill returns the content of a skill by name
func LoadSkill(name string) string {
	skill := SkillRegistry[name]
	if skill == nil {
		var keys []string
		for k := range SkillRegistry {
			keys = append(keys, k)
		}
		return fmt.Sprintf("Skill not found: %s. Available: %s", name, strings.Join(keys, ", "))
	}
	return skill["content"].(string)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func writeTranscript(messages []interface{}) string {
	_ = os.MkdirAll(transcriptDir, 0755)
	path := filepath.Join(transcriptDir, fmt.Sprintf("transcript_%d.jsonl", time.Now().Unix()))
	data, _ := json.Marshal(messages)
	_ = ioutil.WriteFile(path, data, 0644)
	return path
}

func persistLargeOutput(toolUseID string, output string) string {
	if len(output) <= PERSIST_THRESHOLD {
		return output
	}
	_ = os.MkdirAll(toolResultsDir, 0755)
	path := filepath.Join(toolResultsDir, toolUseID+".txt")
	if !fileExists(path) {
		_ = ioutil.WriteFile(path, []byte(output), 0644)
	}
	return fmt.Sprintf("<persisted-output>\nFull output: %s\nPreview:\n%s\n</persisted-output>", path, output[:2000])
}
