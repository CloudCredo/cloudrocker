package models

import (
	"fmt"
	"strconv"
	"strings"
)

type LRPIdentifier struct {
	ProcessGuid  string
	Index        int
	InstanceGuid string
}

func LRPIdentifierFromOpaqueID(opaqueID string) (LRPIdentifier, error) {
	substrings := strings.Split(opaqueID, ".")
	if len(substrings) != 3 {
		return LRPIdentifier{}, fmt.Errorf("invalid opaqueID for LRP: %s", opaqueID)
	}

	index, err := strconv.Atoi(substrings[1])
	if err != nil {
		return LRPIdentifier{}, err
	}

	return LRPIdentifier{
		ProcessGuid:  substrings[0],
		Index:        index,
		InstanceGuid: substrings[2],
	}, nil
}

func (ids LRPIdentifier) OpaqueID() string {
	return fmt.Sprintf("%s.%d.%s", ids.ProcessGuid, ids.Index, ids.InstanceGuid)
}
