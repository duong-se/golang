// TaskManager is a basic task manager.
type TaskManager struct {
	current string
}

func NewTaskManager() *TaskManager {
	return &TaskManager{}
}

func (tm *TaskManager) ScanUnclaimed() []*Task {
	// For now, just return empty slice
	return []*Task{}
}