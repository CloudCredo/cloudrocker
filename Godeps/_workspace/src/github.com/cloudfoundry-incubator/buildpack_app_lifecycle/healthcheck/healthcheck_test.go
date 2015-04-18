package main_test

import (
	"net"
	"os/exec"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gbytes"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gexec"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/ghttp"
)

var _ = Describe("HealthCheck", func() {
	var (
		server     *ghttp.Server
		serverAddr string
	)

	BeforeEach(func() {
		ip := getNonLoopbackIP()
		server = ghttp.NewUnstartedServer()
		listener, err := net.Listen("tcp", ip+":0")
		Ω(err).ShouldNot(HaveOccurred())

		server.HTTPTestServer.Listener = listener
		serverAddr = listener.Addr().String()
		server.Start()
	})

	runHealthCheck := func() *gexec.Session {
		_, port, err := net.SplitHostPort(serverAddr)
		Ω(err).ShouldNot(HaveOccurred())
		session, err := gexec.Start(exec.Command(healthCheck, "-port", port, "-timeout", "100ms"), GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
		return session
	}

	Context("when the address is listening", func() {
		It("exits 0 and logs it passed", func() {
			session := runHealthCheck()
			Eventually(session).Should(gexec.Exit(0))
			Ω(session.Out).Should(gbytes.Say("healthcheck passed"))
		})
	})

	Context("when the address is not listening", func() {
		BeforeEach(func() {
			server.Close()
		})

		It("exits 1 and logs it failed", func() {
			session := runHealthCheck()
			Eventually(session).Should(gexec.Exit(1))
			Ω(session.Out).Should(gbytes.Say("healthcheck failed"))
		})
	})
})

func getNonLoopbackIP() string {
	interfaces, err := net.Interfaces()
	Ω(err).ShouldNot(HaveOccurred())
	for _, intf := range interfaces {
		addrs, err := intf.Addrs()
		if err != nil {
			continue
		}

		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	Fail("no non-loopback address found")
	panic("non-reachable")
}
