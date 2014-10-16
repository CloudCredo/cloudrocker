package models

import "encoding/json"

type DesiredLRP struct {
	ProcessGuid string           `json:"process_guid"`
	Domain      string           `json:"domain"`
	RootFSPath  string           `json:"root_fs"`
	Instances   int              `json:"instances"`
	Stack       string           `json:"stack"`
	Actions     []ExecutorAction `json:"actions"`
	DiskMB      int              `json:"disk_mb"`
	MemoryMB    int              `json:"memory_mb"`
	Ports       []PortMapping    `json:"ports"`
	Routes      []string         `json:"routes"`
	Log         LogConfig        `json:"log"`
}

type DesiredLRPChange struct {
	Before *DesiredLRP
	After  *DesiredLRP
}

func NewDesiredLRPFromJSON(payload []byte) (DesiredLRP, error) {
	var lrp DesiredLRP

	err := json.Unmarshal(payload, &lrp)
	if err != nil {
		return DesiredLRP{}, err
	}

	if lrp.Domain == "" {
		return DesiredLRP{}, ErrInvalidJSONMessage{"domain"}
	}

	if lrp.ProcessGuid == "" {
		return DesiredLRP{}, ErrInvalidJSONMessage{"process_guid"}
	}

	if lrp.Stack == "" {
		return DesiredLRP{}, ErrInvalidJSONMessage{"stack"}
	}

	if len(lrp.Actions) == 0 {
		return DesiredLRP{}, ErrInvalidJSONMessage{"actions"}
	}

	return lrp, nil
}

func (desired DesiredLRP) ToJSON() []byte {
	bytes, err := json.Marshal(desired)
	if err != nil {
		panic(err)
	}

	return bytes
}
