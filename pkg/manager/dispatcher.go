package manager

import (
	"fmt"
	"sync"
	"time"

	"rangeforge-ue/pkg/models"
	"github.com/google/uuid"
)

// TaskDispatcher orchestrates task distribution and tracks execution results.
type TaskDispatcher struct {
	mu               sync.Mutex
	taskQueues       map[string][]models.EmulationTask
	taskResults      map[string]models.TaskResult
	personaOverrides map[string]string
}

// NewTaskDispatcher constructs a task dispatcher.
func NewTaskDispatcher() *TaskDispatcher {
	return &TaskDispatcher{
		taskQueues:       make(map[string][]models.EmulationTask),
		taskResults:      make(map[string]models.TaskResult),
		personaOverrides: make(map[string]string),
	}
}

// DispatchCommand enqueues an execution instruction for a target agent.
func (d *TaskDispatcher) DispatchCommand(req models.CommandRequest) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	taskID := fmt.Sprintf("cmd-%s", uuid.New().String()[:8])
	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = 30
	}

	task := models.EmulationTask{
		TaskID: taskID,
		Type:   models.TaskTypeCustomCommand,
		Parameters: map[string]interface{}{
			"command": req.Command,
			"shell":   req.Shell,
		},
		CreatedAt:      time.Now().UTC(),
		TimeoutSeconds: timeout,
	}

	d.taskQueues[req.TargetAgentID] = append(d.taskQueues[req.TargetAgentID], task)
	return taskID, nil
}

// DispatchBatchCommand enqueues an execution instruction for multiple target agents atomically.
func (d *TaskDispatcher) DispatchBatchCommand(targetIDs []string, req models.CommandRequest) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = 30
	}

	taskIDs := make([]string, 0, len(targetIDs))
	for _, targetID := range targetIDs {
		if targetID == "" {
			continue
		}
		taskID := fmt.Sprintf("cmd-%s", uuid.New().String()[:8])
		task := models.EmulationTask{
			TaskID: taskID,
			Type:   models.TaskTypeCustomCommand,
			Parameters: map[string]interface{}{
				"command": req.Command,
				"shell":   req.Shell,
			},
			CreatedAt:      time.Now().UTC(),
			TimeoutSeconds: timeout,
		}
		d.taskQueues[targetID] = append(d.taskQueues[targetID], task)
		taskIDs = append(taskIDs, taskID)
	}
	return taskIDs, nil
}

// FetchTasks retrieves and drains queued tasks for an agent.
func (d *TaskDispatcher) FetchTasks(agentID string) []models.EmulationTask {
	d.mu.Lock()
	defer d.mu.Unlock()

	tasks, ok := d.taskQueues[agentID]
	if !ok || len(tasks) == 0 {
		return nil
	}

	// Drain queue
	d.taskQueues[agentID] = nil
	return tasks
}

// RecordResult stores the result of a completed task.
func (d *TaskDispatcher) RecordResult(result models.TaskResult) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Bound memory: prune oldest results if exceeding 500
	if len(d.taskResults) >= 500 {
		var oldestID string
		var oldestTime time.Time
		for id, r := range d.taskResults {
			if oldestID == "" || r.CompletedAt.Before(oldestTime) {
				oldestID = id
				oldestTime = r.CompletedAt
			}
		}
		if oldestID != "" {
			delete(d.taskResults, oldestID)
		}
	}
	d.taskResults[result.TaskID] = result
}

// GetResult retrieves execution result by task ID.
func (d *TaskDispatcher) GetResult(taskID string) (models.TaskResult, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	res, ok := d.taskResults[taskID]
	return res, ok
}

// SetPersona overrides the active persona for an agent.
func (d *TaskDispatcher) SetPersona(agentID string, persona string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.personaOverrides[agentID] = persona
}

// GetPersona returns persona override if specified.
func (d *TaskDispatcher) GetPersona(agentID string) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.personaOverrides[agentID]
}
