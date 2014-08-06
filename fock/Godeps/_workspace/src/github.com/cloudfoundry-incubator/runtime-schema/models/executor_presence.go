package models

import "encoding/json"

type ExecutorPresence struct {
	ExecutorID string `json:"executor_id"`
	Stack      string `json:"stack"`
}

func NewExecutorPresenceFromJSON(payload []byte) (ExecutorPresence, error) {
	var task ExecutorPresence

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return ExecutorPresence{}, err
	}

	return task, nil
}

func (presence ExecutorPresence) ToJSON() []byte {
	bytes, err := json.Marshal(presence)
	if err != nil {
		panic(err)
	}

	return bytes
}
