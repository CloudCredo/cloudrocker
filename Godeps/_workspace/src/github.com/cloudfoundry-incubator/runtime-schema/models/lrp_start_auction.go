package models

import "encoding/json"

type LRPStartAuctionState int

const (
	LRPStartAuctionStateInvalid LRPStartAuctionState = iota
	LRPStartAuctionStatePending
	LRPStartAuctionStateClaimed
)

type LRPStartAuction struct {
	DesiredLRP DesiredLRP `json:"desired_lrp"`

	InstanceGuid string `json:"instance_guid"`
	Index        int    `json:"index"`

	State     LRPStartAuctionState `json:"state"`
	UpdatedAt int64                `json:"updated_at"`
}

func NewLRPStartAuctionFromJSON(payload []byte) (LRPStartAuction, error) {
	var task LRPStartAuction

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return LRPStartAuction{}, err
	}

	if task.InstanceGuid == "" {
		return LRPStartAuction{}, ErrInvalidJSONMessage{"instance_guid"}
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
		ProcessGuid:  auction.DesiredLRP.ProcessGuid,
		Index:        auction.Index,
		InstanceGuid: auction.InstanceGuid,
	}
}
