package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

var network = flag.String(
	"network",
	"tcp",
	"network type to dial with (e.g. unix, tcp)",
)

var port = flag.String(
	"port",
	"8080",
	"port to test",
)

var timeout = flag.Duration(
	"timeout",
	1*time.Second,
	"dial timeout",
)

func main() {
	flag.Parse()

	interfaces, err := net.Interfaces()
	if err == nil {
		for _, intf := range interfaces {
			addrs, err := intf.Addrs()
			if err != nil {
				continue
			}
			for _, a := range addrs {
				if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						addr := ipnet.IP.String() + ":" + *port
						conn, err := net.DialTimeout(*network, addr, *timeout)
						if err == nil {
							conn.Close()
							fmt.Println("healthcheck passed")
							os.Exit(0)
						}
					}
				}
			}
		}
	}

	fmt.Println("healthcheck failed")
	os.Exit(1)
}
