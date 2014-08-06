package models

import "encoding/json"

type ActualLRPState int

const (
	ActualLRPStateInvalid ActualLRPState = iota
	ActualLRPStateStarting
	ActualLRPStateRunning
)

type ActualLRPChange struct {
	Before *ActualLRP
	After  *ActualLRP
}

type ActualLRP struct {
	ProcessGuid  string `json:"process_guid"`
	InstanceGuid string `json:"instance_guid"`
	ExecutorID   string `json:"executor_id"`

	Index int `json:"index"`

	Host  string        `json:"host"`
	Ports []PortMapping `json:"ports"`

	State ActualLRPState `json:"state"`
	Since int64          `json:"since"`
}

func NewActualLRPFromJSON(payload []byte) (ActualLRP, error) {
	var task ActualLRP

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return ActualLRP{}, err
	}

	if task.ProcessGuid == "" {
		return ActualLRP{}, ErrInvalidJSONMessage{"process_guid"}
	}

	if task.InstanceGuid == "" {
		return ActualLRP{}, ErrInvalidJSONMessage{"instance_guid"}
	}

	if task.ExecutorID == "" {
		return ActualLRP{}, ErrInvalidJSONMessage{"executor_id"}
	}

	return task, nil
}

func (actual ActualLRP) ToJSON() []byte {
	bytes, err := json.Marshal(actual)
	if err != nil {
		panic(err)
	}

	return bytes
}
