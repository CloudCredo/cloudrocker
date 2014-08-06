package models

type LogConfig struct {
	Guid       string `json:"guid"`
	SourceName string `json:"source_name"`
	Index      *int   `json:"index"`
}
