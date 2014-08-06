package models

import "encoding/json"

type DesiredLRP struct {
	// required
	ProcessGuid string `json:"process_guid"`
	Source      string `json:"source"`
	Stack       string `json:"stack"`

	// optional
	Instances       int                   `json:"instances"`
	MemoryMB        int                   `json:"memory_mb"`
	DiskMB          int                   `json:"disk_mb"`
	FileDescriptors uint64                `json:"file_descriptors"`
	StartCommand    string                `json:"start_command"`
	Environment     []EnvironmentVariable `json:"environment"`
	Routes          []string              `json:"routes"`
	LogGuid         string                `json:"log_guid"`
}

type DesiredLRPChange struct {
	Before *DesiredLRP
	After  *DesiredLRP
}

func NewDesiredLRPFromJSON(payload []byte) (DesiredLRP, error) {
	var task DesiredLRP

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return DesiredLRP{}, err
	}

	if task.ProcessGuid == "" {
		return DesiredLRP{}, ErrInvalidJSONMessage{"process_guid"}
	}

	if task.Source == "" {
		return DesiredLRP{}, ErrInvalidJSONMessage{"source"}
	}

	if task.Stack == "" {
		return DesiredLRP{}, ErrInvalidJSONMessage{"stack"}
	}

	return task, nil
}

func (desired DesiredLRP) ToJSON() []byte {
	bytes, err := json.Marshal(desired)
	if err != nil {
		panic(err)
	}

	return bytes
}
