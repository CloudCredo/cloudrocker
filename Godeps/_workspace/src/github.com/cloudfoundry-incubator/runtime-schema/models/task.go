package models

import (
	"encoding/json"
)

type TaskState int

const (
	TaskStateInvalid TaskState = iota
	TaskStatePending
	TaskStateClaimed
	TaskStateRunning
	TaskStateCompleted
	TaskStateResolving
)

type TaskType string

const (
	TaskTypeStaging          TaskType = "Staging"
	TaskTypeDropletMigration TaskType = "DropletMigration"
)

type Task struct {
	Guid       string           `json:"guid"`
	Actions    []ExecutorAction `json:"actions"`
	Stack      string           `json:"stack"`
	MemoryMB   int              `json:"memory_mb"`
	DiskMB     int              `json:"disk_mb"`
	CpuPercent float64          `json:"cpu_percent"`
	Log        LogConfig        `json:"log"`
	CreatedAt  int64            `json:"created_at"` //  the number of nanoseconds elapsed since January 1, 1970 UTC
	UpdatedAt  int64            `json:"updated_at"`

	Type TaskType `json:"type"`

	State TaskState `json:"state"`

	ExecutorID string `json:"executor_id"`

	ContainerHandle string `json:"container_handle"`

	Result        string `json:"result"`
	Failed        bool   `json:"failed"`
	FailureReason string `json:"failure_reason"`

	Annotation string `json:"annotation,omitempty"`
}

func NewTaskFromJSON(payload []byte) (Task, error) {
	var task Task

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return Task{}, err
	}

	if task.Guid == "" {
		return Task{}, ErrInvalidJSONMessage{"guid"}
	}

	if len(task.Actions) == 0 {
		return Task{}, ErrInvalidJSONMessage{"actions"}
	}

	if task.Stack == "" {
		return Task{}, ErrInvalidJSONMessage{"stack"}
	}

	return task, nil
}

func (task Task) ToJSON() []byte {
	bytes, err := json.Marshal(task)
	if err != nil {
		panic(err)
	}

	return bytes
}
