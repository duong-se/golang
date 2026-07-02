package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/liushuangls/go-anthropic/v2" // Thư viện Anthropic SDK cho Go
)

// ── Các Hằng Số và Biến Toàn Cục ──
const (
	DefaultMaxTokens      = 8000
	EscalatedMaxTokens    = 16000
	MaxRetries            = 3
	MaxConsecutive529     = 2
	MaxRecoveryRetries    = 2
	BaseDelayMs           = 500
	ContextLimit          = 50000
	KeepRecentToolResults = 3
	PersistThreshold      = 30000
	ContinuationPrompt    = "Continue from the previous response. Do not repeat completed work."
	PromptColor           = "\033[36ms20 >> \033[0m"
)

var (
	WorkDir        string
	Client         *anthropic.Client
	ModelID        string
	PrimaryModel   string
	FallbackModel  string
	SkillsDir      string
	TranscriptDir  string
	ToolResultsDir string
	TasksDir       string
	WorktreesDir   string
	MailboxDir     string
	DurablePath    string
	MemoryIndex    string
	MemoryDir      string

	CurrentTodos    []map[string]interface{}
	CurrentTodosMu  sync.Mutex
	CliActive       = false
	RoundsSinceTodo = 0
	AgentLock       sync.Mutex
	ValidWtName     = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)
)

func initEnv() {
	_ = godotenv.Load(".env")
	dir, _ := os.Getwd()
	WorkDir = dir

	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	Client = anthropic.NewClient(apiKey, anthropic.WithBaseURL(baseURL))

	ModelID = os.Getenv("MODEL_ID")
	PrimaryModel = ModelID
	FallbackModel = os.Getenv("FALLBACK_MODEL_ID")

	SkillsDir = filepath.Join(WorkDir, "skills")
	TranscriptDir = filepath.Join(WorkDir, ".transcripts")
	ToolResultsDir = filepath.Join(WorkDir, ".task_outputs", "tool-results")
	TasksDir = filepath.Join(WorkDir, ".tasks")
	WorktreesDir = filepath.Join(WorkDir, ".worktrees")
	MailboxDir = filepath.Join(WorkDir, ".mailboxes")
	DurablePath = filepath.Join(WorkDir, ".scheduled_tasks.json")
	MemoryDir = filepath.Join(WorkDir, ".memory")
	MemoryIndex = filepath.Join(MemoryDir, "MEMORY.md")

	_ = os.MkdirAll(TasksDir, 0755)
	_ = os.MkdirAll(WorktreesDir, 0755)
	_ = os.MkdirAll(MailboxDir, 0755)
}

func terminalPrint(text string) {
	// Ở bản Go đơn giản hóa phần readline của Python thành in thẳng ra terminal
	fmt.Println(text)
}

// ── Task System ──
type Task struct {
	ID          string   `json:"id"`
	Subject     string   `json:"subject"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Owner       *string  `json:"owner"`
	BlockedBy   []string `json:"blockedBy"`
	Worktree    *string  `json:"worktree"`
}

func taskPath(taskID string) string {
	return filepath.Join(TasksDir, taskID+".json")
}

func createTask(subject string, description string, blockedBy []string) Task {
	if blockedBy == nil {
		blockedBy = []string{}
	}
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("task_%d_%04d", time.Now().Unix(), rand.Intn(10000))
	task := Task{
		ID:          id,
		Subject:     subject,
		Description: description,
		Status:      "pending",
		Owner:       nil,
		BlockedBy:   blockedBy,
		Worktree:    nil,
	}
	saveTask(task)
	return task
}

func saveTask(task Task) {
	data, _ := json.MarshalIndent(task, "", "  ")
	_ = os.WriteFile(taskPath(task.ID), data, 0644)
}

func loadTask(taskID string) (Task, error) {
	var task Task
	data, err := os.ReadFile(taskPath(taskID))
	if err != nil {
		return task, err
	}
	err = json.Unmarshal(data, &task)
	return task, err
}

func listTasks() []Task {
	files, _ := filepath.Glob(filepath.Join(TasksDir, "task_*.json"))
	sort.Strings(files)
	var tasks []Task
	for _, f := range files {
		var t Task
		data, _ := os.ReadFile(f)
		_ = json.Unmarshal(data, &t)
		tasks = append(tasks, t)
	}
	return tasks
}

func canStart(taskID string) bool {
	task, err := loadTask(taskID)
	if err != nil {
		return false
	}
	for _, depID := range task.BlockedBy {
		if _, err := os.Stat(taskPath(depID)); os.IsNotExist(err) {
			return false
		}
		dep, _ := loadTask(depID)
		if dep.Status != "completed" {
			return false
		}
	}
	return true
}

func claimTask(taskID string, owner string) string {
	task, err := loadTask(taskID)
	if err != nil {
		return fmt.Sprintf("Task %s not found", taskID)
	}
	if task.Status != "pending" {
		return fmt.Sprintf("Task %s is %s, cannot claim", taskID, task.Status)
	}
	if task.Owner != nil {
		return fmt.Sprintf("Task %s already owned by %s", taskID, *task.Owner)
	}
	if !canStart(taskID) {
		return "Cannot start — blocked by dependencies"
	}
	task.Owner = &owner
	task.Status = "in_progress"
	saveTask(task)
	fmt.Printf("  \033[36m[claim] %s → in_progress\033[0m\n", task.Subject)
	return fmt.Sprintf("Claimed %s (%s)", task.ID, task.Subject)
}

func completeTask(taskID string) string {
	task, err := loadTask(taskID)
	if err != nil {
		return fmt.Sprintf("Task %s not found", taskID)
	}
	if task.Status != "in_progress" {
		return fmt.Sprintf("Task %s is %s, cannot complete", taskID, task.Status)
	}
	task.Status = "completed"
	saveTask(task)
	fmt.Printf("  \033[32m[complete] %s ✓\033[0m\n", task.Subject)
	return fmt.Sprintf("Completed %s (%s)", task.ID, task.Subject)
}

// ── Worktree System ──
func runGit(args []string) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = WorkDir
	out, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(out))
	if len(outputStr) > 5000 {
		outputStr = outputStr[:5000]
	}
	if err != nil {
		return false, outputStr
	}
	return true, outputStr
}

func createWorktree(name string, taskID string) string {
	if !ValidWtName.MatchString(name) {
		return "Error: Invalid worktree name"
	}
	path := filepath.Join(WorktreesDir, name)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Sprintf("Worktree '%s' already exists", name)
	}
	ok, res := runGit([]string{"worktree", "add", path, "-b", "wt/" + name, "HEAD"})
	if !ok {
		return "Git error: " + res
	}
	if taskID != "" {
		t, err := loadTask(taskID)
		if err == nil {
			t.Worktree = &name
			saveTask(t)
		}
	}
	return fmt.Sprintf("Worktree '%s' created at %s", name, path)
}

// ── Skill System ──
var SkillRegistry = make(map[string]map[string]string)

func scanSkills() {
	files, _ := os.ReadDir(SkillsDir)
	for _, f := range files {
		if f.IsDir() {
			manifest := filepath.Join(SkillsDir, f.Name(), "SKILL.md")
			if data, err := os.ReadFile(manifest); err == nil {
				SkillRegistry[f.Name()] = map[string]string{
					"name":        f.Name(),
					"description": "Skill " + f.Name(),
					"content":     string(data),
				}
			}
		}
	}
}

// ── MessageBus ──
type Message struct {
	From     string                 `json:"from"`
	To       string                 `json:"to"`
	Content  string                 `json:"content"`
	Type     string                 `json:"type"`
	Ts       float64                `json:"ts"`
	Metadata map[string]interface{} `json:"metadata"`
}

type MessageBus struct {
	mu sync.Mutex
}

var BUS MessageBus

func (b *MessageBus) Send(from, to, content, msgType string, metadata map[string]interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	msg := Message{
		From:     from,
		To:       to,
		Content:  content,
		Type:     msgType,
		Ts:       float64(time.Now().Unix()),
		Metadata: metadata,
	}
	file, _ := os.OpenFile(filepath.Join(MailboxDir, to+".jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer file.Close()
	data, _ := json.Marshal(msg)
	_, _ = file.Write(append(data, '\n'))
}

func (b *MessageBus) ReadInbox(agent string) []Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	path := filepath.Join(MailboxDir, agent+".jsonl")
	file, err := os.Open(path)
	if err != nil {
		return []Message{}
	}
	defer file.Close()

	var msgs []Message
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal([]byte(scanner.Text()), &msg); err == nil {
			msgs = append(msgs, msg)
		}
	}
	file.Close()
	_ = os.Remove(path)
	return msgs
}

// ── Hooks ──
type HookFunc func(block interface{}) interface{}

var Hooks = make(map[string][]HookFunc)

func registerHook(event string, f HookFunc) {
	Hooks[event] = append(Hooks[event], f)
}

// ── Background Tasks ──
var (
	bgCounter         = 0
	backgroundTasks   = make(map[string]map[string]interface{})
	backgroundResults = make(map[string]string)
	backgroundLock    sync.Mutex
)

// ── Cron Scheduler ──
type CronJob struct {
	ID        string `json:"id"`
	Cron      string `json:"cron"`
	Prompt    string `json:"prompt"`
	Recurring bool   `json:"recurring"`
	Durable   bool   `json:"durable"`
}

var (
	scheduledJobs = make(map[string]CronJob)
	cronQueue     []CronJob
	cronLock      sync.Mutex
)

func cronSchedulerLoop() {
	for {
		time.Sleep(1 * time.Second)
		// Thực hiện logic khớp giờ cron ở đây...
	}
}

// ── Thao tác File cơ bản làm Tools cho Agent ──
func runBash(command string, dir string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	if dir != "" {
		cmd.Dir = dir
	} else {
		cmd.Dir = WorkDir
	}
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func consumeCronQueue() []CronJob {
	cronLock.Lock()
	defer cronLock.Unlock() // Tự động unlock khi hàm return (tương đương với `with` trong Python)

	// Nếu queue rỗng, trả về nil ngay để tiết kiệm hiệu năng
	if len(cronQueue) == 0 {
		return nil
	}

	// Sao chép (slice) toàn bộ phần tử hiện tại sang mảng 'fired'
	fired := make([]CronJob, len(cronQueue))
	copy(fired, cronQueue)

	// Xóa sạch hàng đợi (tương đương với cron_queue.clear())
	cronQueue = []CronJob{}

	return fired
}

func collectBackgroundResults() []string {
	// Khóa lock một lần duy nhất cho toàn bộ quá trình đọc/ghi map để tối ưu hiệu năng
	backgroundLock.Lock()

	// Khai báo danh sách các ID đã sẵn sàng
	var ready []string
	for bgID, task := range backgroundTasks {
		if status, ok := task["status"].(string); ok && status == "completed" {
			ready = append(ready, bgID)
		}
	}

	// Nếu không có task nào xong, unlock và return sớm
	if len(ready) == 0 {
		backgroundLock.Unlock()
		return nil
	}

	// Struct tạm để giữ dữ liệu cần thiết phục vụ cho việc build chuỗi thông báo sau khi unlock
	type TaskSnapshot struct {
		ID      string
		Command string
		Output  string
	}
	snapshots := make([]TaskSnapshot, 0, len(ready))

	// Trích xuất dữ liệu ra mảng tạm và xóa luôn khỏi map (tương đương với .pop() trong Python)
	for _, bgID := range ready {
		task := backgroundTasks[bgID]
		delete(backgroundTasks, bgID) // pop task

		output := backgroundResults[bgID]
		delete(backgroundResults, bgID) // pop result

		command, _ := task["command"].(string)

		snapshots = append(snapshots, TaskSnapshot{
			ID:      bgID,
			Command: command,
			Output:  output,
		})
	}

	// Giải phóng lock ngay khi thao tác xong trên map, việc dựng chuỗi HTML phía sau không cần lock nữa
	backgroundLock.Unlock()

	// Tiến hành build danh sách notifications
	var notifications []string
	for _, snap := range snapshots {
		// Xử lý cắt chuỗi tương đương output[:200] trong Python
		summary := snap.Output
		if len(summary) > 200 {
			summary = summary[:200]
		}

		// Sử dụng cú pháp raw string literal (dấu ` `) của Go để viết chuỗi nhiều dòng sạch đẹp
		notification := fmt.Sprintf("<task_notification>\n"+
			"  <task_id>%s</task_id>\n"+
			"  <status>completed</status>\n"+
			"  <command>%s</command>\n"+
			"  <summary>%s</summary>\n"+
			"</task_notification>",
			snap.ID, snap.Command, summary)

		notifications = append(notifications, notification)
	}

	return notifications
}

func injectBackgroundNotifications(messages *[]anthropic.Message) {
	notes := collectBackgroundResults()

	// Nếu có thông báo mới (tương đương với `if notes:` trong Python)
	if len(notes) > 0 {
		var contentBlocks []anthropic.MessageContent

		// List comprehension của Python được chuyển thành vòng lặp for trong Go
		for _, note := range notes {
			contentBlocks = append(contentBlocks, anthropic.NewTextMessageContent(note))
		}

		// Append block nội dung mới vào danh sách tin nhắn thông qua con trỏ
		*messages = append(*messages, anthropic.Message{
			Role:    anthropic.RoleUser,
			Content: contentBlocks,
		})
	}
}

// Hàm bổ trợ kiểm tra loại tin nhắn hoặc loại block
func isToolResultMessage(msg anthropic.Message) bool {
	if msg.Role != anthropic.RoleUser {
		return false
	}
	for _, b := range msg.Content {
		if b.Type == anthropic.MessagesContentTypeToolResult {
			return true
		}
	}
	return false
}

func messageHasToolUse(msg anthropic.Message) bool {
	if msg.Role != anthropic.RoleAssistant {
		return false
	}
	for _, b := range msg.Content {
		if b.Type == anthropic.MessagesContentTypeToolUse {
			return true
		}
	}
	return false
}

func toolResultBudget(messages []map[string]interface{}, maxBytes ...int) []map[string]interface{} {
	if len(messages) == 0 {
		return messages
	}

	// Đặt giá trị mặc định cho max_bytes = 200,000 giống Python
	limit := 200000
	if len(maxBytes) > 0 {
		limit = maxBytes[0]
	}

	// Lấy phần tử cuối cùng
	last := messages[len(messages)-1]
	role, _ := last["role"].(string)
	content, isList := last["content"].([]interface{})

	// Nếu không phải role 'user' hoặc content không phải là một list các block -> Trả về luôn
	if role != "user" || !isList {
		return messages
	}

	// Định nghĩa struct tạm để lưu index và block phục vụ việc tính toán/sắp xếp
	type BlockPair struct {
		Index int
		Block map[string]interface{}
		Size  int
	}

	var blocks []BlockPair

	// Hàm bổ trợ tính tổng dung lượng hiện tại của các tool_result
	calcTotal := func() int {
		sum := 0
		for _, b := range content {
			if bMap, ok := b.(map[string]interface{}); ok {
				if bMap["type"] == "tool_result" {
					contentStr := fmt.Sprintf("%v", bMap["content"])
					sum += len(contentStr)
				}
			}
		}
		return sum
	}

	// Lọc và thu thập các block thuộc loại "tool_result"
	for i, b := range content {
		if bMap, ok := b.(map[string]interface{}); ok {
			if bMap["type"] == "tool_result" {
				contentStr := fmt.Sprintf("%v", bMap["content"])
				blocks = append(blocks, BlockPair{
					Index: i,
					Block: bMap,
					Size:  len(contentStr),
				})
			}
		}
	}

	total := calcTotal()
	if total <= limit {
		return messages
	}

	// Sắp xếp các block theo độ dài content giảm dần (reverse=True)
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Size > blocks[j].Size
	})

	// Vòng lặp nén các block lớn nhất
	for _, pair := range blocks {
		if total <= limit {
			break
		}

		text := fmt.Sprintf("%v", pair.Block["content"])
		toolUseID, _ := pair.Block["tool_use_id"].(string)
		if toolUseID == "" {
			toolUseID = "unknown"
		}

		// Thay thế content cũ bằng đường dẫn lưu file mới thông qua persistLargeOutput
		pair.Block["content"] = persistLargeOutput(toolUseID, text)

		// Tính toán lại tổng dung lượng sau khi nén
		total = calcTotal()
	}

	return messages
}

func prepareContext(messages *[]anthropic.Message) []anthropic.Message {
	if messages == nil {
		return nil
	}

	// Giải bọc con trỏ để thao tác và cập nhật lại slice gốc qua từng pipeline
	*messages = toolResultBudget(*messages)
	*messages = snipCompact(*messages)
	*messages = microCompact(*messages)

	// Kiểm tra nếu kích thước vượt ngưỡng CONTEXT_LIMIT thì tiến hành nén lịch sử
	if estimateSize(*messages) > ContextLimit {
		*messages = compactHistory(*messages)
	}

	// Trả về slice đã được xử lý (giống như `return messages` trong Python)
	return *messages
}

func agentLoop(messages *[]anthropic.Message, contextData map[string]interface{}) {
	// Khóa các biến global liên quan đến todo (tránh race condition trong Go)
	CurrentTodosMu.Lock()
	roundsSinceTodoLocal := RoundsSinceTodo
	CurrentTodosMu.Unlock()

	state := &RecoveryState{}
	maxTokens := DefaultMaxTokens

	for {
		// 1. Tiêu thụ hàng đợi cron
		fired := consumeCronQueue()
		for _, job := range fired {
			*messages = append(*messages, anthropic.Message{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewTextMessageContent(fmt.Sprintf("[Scheduled] %s", job.Prompt)),
				},
			})

			// Giới hạn chuỗi in ra tương tự job.prompt[:60] trong Python
			limit := 60
			if len(job.Prompt) < limit {
				limit = len(job.Prompt)
			}
			fmt.Printf("  \033[35m[cron inject] %s\033[0m\n", job.Prompt[:limit])
		}

		// 2. Inject background notifications
		injectBackgroundNotifications(messages)

		// 3. Kiểm tra nhắc nhở cập nhật TODO
		if roundsSinceTodoLocal >= 3 {
			*messages = append(*messages, anthropic.Message{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewTextMessageContent("<reminder>Update your todos.</reminder>"),
				},
			})
			roundsSinceTodoLocal = 0
		}

		// 4. Chuẩn bị ngữ cảnh và cập nhật
		prepareContext(messages)
		contextData = updateContext(contextData, *messages)
		tools, handlers := assembleToolPool()

		// 5. Gọi LLM kèm xử lý lỗi (Prompt Too Long)
		response, err := callLLM(*messages, contextData, tools, state, maxTokens)
		if err != nil {
			if isPromptTooLongError(err) && !state.HasAttemptedReactiveCompact {
				*messages = reactiveCompact(*messages)
				state.HasAttemptedReactiveCompact = true
				continue
			}

			// Ghi nhận lỗi hệ thống vào hội thoại dưới dạng block text của assistant
			errText := fmt.Sprintf("[Error] %s: %s", reflect.TypeOf(err).Name(), err.Error())
			*messages = append(*messages, anthropic.Message{
				Role: anthropic.RoleAssistant,
				Content: []anthropic.MessageContent{
					anthropic.NewTextMessageContent(errText),
				},
			})
			return
		}

		// 6. Xử lý khi vượt ngưỡng max_tokens (Stop Reason)
		if response.StopReason == anthropic.StopReasonMaxTokens {
			if !state.HasEscalated {
				maxTokens = EscalatedMaxTokens
				state.HasEscalated = true
				fmt.Printf("  \033[33m[max_tokens] retry with %d\033[0m\n", maxTokens)
				continue
			}

			*messages = append(*messages, anthropic.Message{
				Role:    anthropic.RoleAssistant,
				Content: response.Content,
			})

			if state.RecoveryCount < MaxRecoveryRetries {
				*messages = append(*messages, anthropic.Message{
					Role: anthropic.RoleUser,
					Content: []anthropic.MessageContent{
						anthropic.NewTextMessageContent(ContinuationPrompt),
					},
				})
				state.RecoveryCount++
				continue
			}
			return
		}

		// Khôi phục lại trạng thái token ban đầu sau khi thành công
		maxTokens = DefaultMaxTokens
		state.HasEscalated = false

		*messages = append(*messages, anthropic.Message{
			Role:    anthropic.RoleAssistant,
			Content: response.Content,
		})

		// 7. Kiểm tra xem LLM có gọi tool nào không
		if !hasToolUse(response.Content) {
			triggerHooks("Stop", messages)
			return
		}

		var results []map[string]interface{} // Chứa kết quả tool để gom nhóm trả về
		compactedNow := false

		// 8. Duyệt qua từng content block để xử lý Tool Call
		for _, block := range response.Content {
			if block.Type != anthropic.MessageContentTypeToolUse {
				continue
			}

			fmt.Printf("\033[36m> %s\033[0m\n", block.Name)

			// Xử lý tool đặc biệt: compact
			if block.Name == "compact" {
				*messages = compactHistory(*messages)
				*messages = append(*messages, anthropic.Message{
					Role: anthropic.RoleUser,
					Content: []anthropic.MessageContent{
						anthropic.NewTextMessageContent("[Compacted. Continue with summarized context.]"),
					},
				})
				compactedNow = true
				break
			}

			// Hook kiểm tra quyền trước khi chạy tool
			blocked := triggerHooks("PreToolUse", block)
			if blocked != nil {
				results = append(results, map[string]interface{}{
					"type":        "tool_result",
					"tool_use_id": block.ID,
					"content":     fmt.Sprintf("%v", blocked),
				})
				continue
			}

			// Kiểm tra và chạy ngầm nếu là tác vụ tốn thời gian (slow task)
			if shouldRunBackground(block.Name, block.Input) {
				bgID := startBackgroundTask(block, handlers)
				output := fmt.Sprintf("[Background task %s started] Result will arrive as a task_notification.", bgID)
				results = append(results, map[string]interface{}{
					"type":        "tool_result",
					"tool_use_id": block.ID,
					"content":     output,
				})
				continue
			}

			// Thực thi Tool thông thường thông qua Handler
			handler := handlers[block.Name]
			output := callToolHandler(handler, block.Input, block.Name)
			triggerHooks("PostToolUse", block, output)

			// In preview kết quả ra màn hình tương tự python: str(output)[:300]
			outputStr := fmt.Sprintf("%v", output)
			previewLen := 300
			if len(outputStr) < previewLen {
				previewLen = len(outputStr)
			}
			fmt.Println(outputStr[:previewLen])

			// Cập nhật bộ đếm vòng lặp todo
			if block.Name == "todo_write" {
				roundsSinceTodoLocal = 0
			} else {
				roundsSinceTodoLocal++
			}

			results = append(results, map[string]interface{}{
				"type":        "tool_result",
				"tool_use_id": block.ID,
				"content":     output,
			})
		}

		if compactedNow {
			continue
		}

		// Đồng bộ ngược lại bộ đếm roundsSinceTodo vào biến global an toàn
		CurrentTodosMu.Lock()
		RoundsSinceTodo = roundsSinceTodoLocal
		CurrentTodosMu.Unlock()

		// Gửi toàn bộ kết quả của các Tool về lại cho LLM (User role)
		*messages = append(*messages, anthropic.Message{
			Role:    anthropic.RoleUser,
			Content: buildUserContent(results),
		})
	}
}

// ── Hàm Main chính ──
func main() {
	initEnv()
	scanSkills()
	CliActive = true

	fmt.Println("Enter a question, press Enter to send. Type q to quit.")

	go cronSchedulerLoop()

	reader := bufio.NewReader(os.Stdin)
	var history []anthropic.Message
	contextData := map[string]interface{}{"memories": ""}

	for {
		fmt.Print(PromptColor)
		query, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		query = strings.TrimSpace(query)
		if query == "q" || query == "exit" || query == "" {
			break
		}

		// Thêm tin nhắn của User vào history
		history = append(history, anthropic.Message{
			Role: anthropic.RoleUser,
			Content: []anthropic.MessageContent{
				anthropic.NewTextMessageContent(query),
			},
		})

		AgentLock.Lock()
		agentLoop(history, contextData)
		AgentLock.Unlock()
	}
}

type RecoveryState struct {
	HasAttemptedReactiveCompact bool
	HasEscalated                bool
	RecoveryCount               int
}
