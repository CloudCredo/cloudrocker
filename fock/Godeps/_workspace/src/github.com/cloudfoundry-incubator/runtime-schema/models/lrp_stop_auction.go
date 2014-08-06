package models

import "encoding/json"

type LRPStopAuctionState int

const (
	LRPStopAuctionStateInvalid LRPStopAuctionState = iota
	LRPStopAuctionStatePending
	LRPStopAuctionStateClaimed
)

type LRPStopAuction struct {
	ProcessGuid string `json:"process_guid"`
	Index       int    `json:"index"`

	State     LRPStopAuctionState `json:"state"`
	UpdatedAt int64               `json:"updated_at"`
}

func NewLRPStopAuctionFromJSON(payload []byte) (LRPStopAuction, error) {
	var task LRPStopAuction

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return LRPStopAuction{}, err
	}

	if task.ProcessGuid == "" {
		return LRPStopAuction{}, ErrInvalidJSONMessage{"process_guid"}
	}

	return task, nil
}

func (auction LRPStopAuction) ToJSON() []byte {
	bytes, err := json.Marshal(auction)
	if err != nil {
		panic(err)
	}

	return bytes
}
