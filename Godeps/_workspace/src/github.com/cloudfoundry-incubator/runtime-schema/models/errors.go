package models

type ErrInvalidJSONMessage struct {
	MissingField string
}

func (err ErrInvalidJSONMessage) Error() string {
	return "JSON has missing/invalid field: " + err.MissingField
}
