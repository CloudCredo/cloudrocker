package models

type PortMapping struct {
	ContainerPort uint32 `json:"container_port"`
	HostPort      uint32 `json:"host_port,omitempty"`
}
