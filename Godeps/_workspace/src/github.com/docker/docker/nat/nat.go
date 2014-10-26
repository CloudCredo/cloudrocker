package nat

// nat is a convenience package for docker's manipulation of strings describing
// network ports.

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/docker/docker/pkg/parsers"
)

const (
	PortSpecTemplate       = "ip:hostPort:containerPort"
	PortSpecTemplateFormat = "ip:hostPort:containerPort | ip::containerPort | hostPort:containerPort | containerPort"
)

type PortBinding struct {
	HostIp   string
	HostPort string
}

type PortMap map[Port][]PortBinding

type PortSet map[Port]struct{}

// 80/tcp
type Port string

func NewPort(proto, port string) Port {
	return Port(fmt.Sprintf("%s/%s", port, proto))
}

func ParsePort(rawPort string) (int, error) {
	port, err := strconv.ParseUint(rawPort, 10, 16)
	if err != nil {
		return 0, err
	}
	return int(port), nil
}

func (p Port) Proto() string {
	parts := strings.Split(string(p), "/")
	if len(parts) == 1 {
		return "tcp"
	}
	return parts[1]
}

func (p Port) Port() string {
	return strings.Split(string(p), "/")[0]
}

func (p Port) Int() int {
	i, err := ParsePort(p.Port())
	if err != nil {
		panic(err)
	}
	return i
}

// Splits a port in the format of proto/port
func SplitProtoPort(rawPort string) (string, string) {
	var port string
	var proto string

	parts := strings.Split(rawPort, "/")

	if len(parts) == 0 || parts[0] == "" { // we have "" or ""/
		port = ""
		proto = ""
	} else { // we have # or #/  or #/...
		port = parts[0]
		if len(parts) > 1 && parts[1] != "" {
			proto = parts[1] // we have #/...
		} else {
			proto = "tcp" // we have # or #/
		}
	}
	return proto, port
}

func validateProto(proto string) bool {
	for _, availableProto := range []string{"tcp", "udp"} {
		if availableProto == proto {
			return true
		}
	}
	return false
}

// We will receive port specs in the format of ip:public:private/proto and these need to be
// parsed in the internal types
func ParsePortSpecs(ports []string) (map[Port]struct{}, map[Port][]PortBinding, error) {
	var (
		exposedPorts = make(map[Port]struct{}, len(ports))
		bindings     = make(map[Port][]PortBinding)
	)

	for _, rawPort := range ports {
		proto := "tcp"

		if i := strings.LastIndex(rawPort, "/"); i != -1 {
			proto = rawPort[i+1:]
			rawPort = rawPort[:i]
		}
		if !strings.Contains(rawPort, ":") {
			rawPort = fmt.Sprintf("::%s", rawPort)
		} else if len(strings.Split(rawPort, ":")) == 2 {
			rawPort = fmt.Sprintf(":%s", rawPort)
		}

		parts, err := parsers.PartParser(PortSpecTemplate, rawPort)
		if err != nil {
			return nil, nil, err
		}

		var (
			containerPort = parts["containerPort"]
			rawIp         = parts["ip"]
			hostPort      = parts["hostPort"]
		)

		if rawIp != "" && net.ParseIP(rawIp) == nil {
			return nil, nil, fmt.Errorf("Invalid ip address: %s", rawIp)
		}
		if containerPort == "" {
			return nil, nil, fmt.Errorf("No port specified: %s<empty>", rawPort)
		}
		if _, err := strconv.ParseUint(containerPort, 10, 16); err != nil {
			return nil, nil, fmt.Errorf("Invalid containerPort: %s", containerPort)
		}
		if _, err := strconv.ParseUint(hostPort, 10, 16); hostPort != "" && err != nil {
			return nil, nil, fmt.Errorf("Invalid hostPort: %s", hostPort)
		}

		if !validateProto(proto) {
			return nil, nil, fmt.Errorf("Invalid proto: %s", proto)
		}

		port := NewPort(proto, containerPort)
		if _, exists := exposedPorts[port]; !exists {
			exposedPorts[port] = struct{}{}
		}

		binding := PortBinding{
			HostIp:   rawIp,
			HostPort: hostPort,
		}
		bslice, exists := bindings[port]
		if !exists {
			bslice = []PortBinding{}
		}
		bindings[port] = append(bslice, binding)
	}
	return exposedPorts, bindings, nil
}
