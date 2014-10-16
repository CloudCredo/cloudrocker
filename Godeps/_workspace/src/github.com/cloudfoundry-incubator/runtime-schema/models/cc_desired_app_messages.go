package models

import "encoding/json"

type DesireAppRequestFromCC struct {
	ProcessGuid     string                `json:"process_guid"`
	DropletUri      string                `json:"droplet_uri"`
	Stack           string                `json:"stack"`
	StartCommand    string                `json:"start_command"`
	Environment     []EnvironmentVariable `json:"environment"`
	MemoryMB        int                   `json:"memory_mb"`
	DiskMB          int                   `json:"disk_mb"`
	FileDescriptors uint64                `json:"file_descriptors"`
	NumInstances    int                   `json:"num_instances"`
	Routes          []string              `json:"routes"`
	LogGuid         string                `json:"log_guid"`
}

func (d DesireAppRequestFromCC) ToJSON() []byte {
	encoded, _ := json.Marshal(d)
	return encoded
}

type CCDesiredStateServerResponse struct {
	Apps        []DesireAppRequestFromCC `json:"apps"`
	CCBulkToken *json.RawMessage         `json:"token"`
}

type CCBulkToken struct {
	Id int `json:"id"`
}
