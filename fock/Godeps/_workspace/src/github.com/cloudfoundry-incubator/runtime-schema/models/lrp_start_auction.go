package models

import "encoding/json"

type LRPStartAuctionState int

const (
	LRPStartAuctionStateInvalid LRPStartAuctionState = iota
	LRPStartAuctionStatePending
	LRPStartAuctionStateClaimed
)

type LRPStartAuction struct {
	ProcessGuid  string           `json:"process_guid"`
	InstanceGuid string           `json:"instance_guid"`
	Stack        string           `json:"stack"`
	Actions      []ExecutorAction `json:"actions"`

	DiskMB   int `json:"disk_mb"`
	MemoryMB int `json:"memory_mb"`

	Log   LogConfig     `json:"log"`
	Ports []PortMapping `json:"ports"`

	Index int `json:"index"`

	State     LRPStartAuctionState `json:"state"`
	UpdatedAt int64                `json:"updated_at"`
}

func NewLRPStartAuctionFromJSON(payload []byte) (LRPStartAuction, error) {
	var task LRPStartAuction

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return LRPStartAuction{}, err
	}

	if task.ProcessGuid == "" {
		return LRPStartAuction{}, ErrInvalidJSONMessage{"process_guid"}
	}

	if task.InstanceGuid == "" {
		return LRPStartAuction{}, ErrInvalidJSONMessage{"instance_guid"}
	}

	if task.Stack == "" {
		return LRPStartAuction{}, ErrInvalidJSONMessage{"stack"}
	}

	if len(task.Actions) == 0 {
		return LRPStartAuction{}, ErrInvalidJSONMessage{"actions"}
	}

	return task, nil
}

func (auction LRPStartAuction) ToJSON() []byte {
	bytes, err := json.Marshal(auction)
	if err != nil {
		panic(err)
	}

	return bytes
}

func (auction LRPStartAuction) LRPIdentifier() LRPIdentifier {
	return LRPIdentifier{
		ProcessGuid:  auction.ProcessGuid,
		Index:        auction.Index,
		InstanceGuid: auction.InstanceGuid,
	}
}
