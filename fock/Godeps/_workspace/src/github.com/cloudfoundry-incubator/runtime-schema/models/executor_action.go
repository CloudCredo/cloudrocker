package models

import (
	"encoding/json"
	"errors"
	"time"
)

var InvalidActionConversion = errors.New("Invalid Action Conversion")

type DownloadAction struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Extract  bool   `json:"extract"`
	CacheKey string `json:"cache_key"`
}

type UploadAction struct {
	To       string `json:"to"`
	From     string `json:"from"`
	Compress bool   `json:"compress"`
}

type RunAction struct {
	Path           string                `json:"path"`
	Args           []string              `json:"args"`
	Env            []EnvironmentVariable `json:"env"`
	Timeout        time.Duration         `json:"timeout"`
	ResourceLimits ResourceLimits        `json:"resource_limits"`
}

type ResourceLimits struct {
	Nofile *uint64 `json:"nofile,omitempty"`
}

type FetchResultAction struct {
	File string `json:"file"`
}

type TryAction struct {
	Action ExecutorAction `json:"action"`
}

type MonitorAction struct {
	Action             ExecutorAction `json:"action"`
	HealthyHook        HealthRequest  `json:"healthy_hook"`
	UnhealthyHook      HealthRequest  `json:"unhealthy_hook"`
	HealthyThreshold   uint           `json:"healthy_threshold"`
	UnhealthyThreshold uint           `json:"unhealthy_threshold"`
}

type HealthRequest struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

type ParallelAction struct {
	Actions []ExecutorAction `json:"actions"`
}

type EmitProgressAction struct {
	Action         ExecutorAction `json:"action"`
	StartMessage   string         `json:"start_message"`
	SuccessMessage string         `json:"success_message"`
	FailureMessage string         `json:"failure_message"`
}

func EmitProgressFor(action ExecutorAction, startMessage string, successMessage string, failureMessage string) ExecutorAction {
	return ExecutorAction{
		EmitProgressAction{
			Action:         action,
			StartMessage:   startMessage,
			SuccessMessage: successMessage,
			FailureMessage: failureMessage,
		},
	}
}

func Try(action ExecutorAction) ExecutorAction {
	return ExecutorAction{
		TryAction{
			Action: action,
		},
	}
}

func Parallel(actions ...ExecutorAction) ExecutorAction {
	return ExecutorAction{
		ParallelAction{
			Actions: actions,
		},
	}
}

type executorActionEnvelope struct {
	Name          string           `json:"action"`
	ActionPayload *json.RawMessage `json:"args"`
}

type ExecutorAction struct {
	Action interface{} `json:"-"`
}

func (a ExecutorAction) MarshalJSON() ([]byte, error) {
	var envelope executorActionEnvelope

	payload, err := json.Marshal(a.Action)

	if err != nil {
		return nil, err
	}

	switch a.Action.(type) {
	case DownloadAction:
		envelope.Name = "download"
	case RunAction:
		envelope.Name = "run"
	case UploadAction:
		envelope.Name = "upload"
	case FetchResultAction:
		envelope.Name = "fetch_result"
	case EmitProgressAction:
		envelope.Name = "emit_progress"
	case TryAction:
		envelope.Name = "try"
	case MonitorAction:
		envelope.Name = "monitor"
	case ParallelAction:
		envelope.Name = "parallel"
	default:
		return nil, InvalidActionConversion
	}

	envelope.ActionPayload = (*json.RawMessage)(&payload)

	return json.Marshal(envelope)
}

func (a *ExecutorAction) UnmarshalJSON(bytes []byte) error {
	var envelope executorActionEnvelope

	err := json.Unmarshal(bytes, &envelope)
	if err != nil {
		return err
	}

	switch envelope.Name {
	case "download":
		action := DownloadAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &action)
		a.Action = action
	case "run":
		action := RunAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &action)
		a.Action = action
	case "upload":
		action := UploadAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &action)
		a.Action = action
	case "fetch_result":
		action := FetchResultAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &action)
		a.Action = action
	case "emit_progress":
		action := EmitProgressAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &action)
		a.Action = action
	case "try":
		action := TryAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &action)
		a.Action = action
	case "monitor":
		action := MonitorAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &action)
		a.Action = action
	case "parallel":
		action := ParallelAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &action)
		a.Action = action
	default:
		err = InvalidActionConversion
	}

	return err
}
