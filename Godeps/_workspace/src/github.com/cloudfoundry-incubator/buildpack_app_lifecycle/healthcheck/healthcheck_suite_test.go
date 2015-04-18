package main_test

import (
	"testing"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gexec"
)

var healthCheck string

func TestBuildpackLifecycleHealthCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildpack-Lifecycle-HealthCheck Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	healthCheckPath, err := gexec.Build("github.com/cloudfoundry-incubator/buildpack_app_lifecycle/healthcheck")
	Ω(err).ShouldNot(HaveOccurred())
	return []byte(healthCheckPath)
}, func(healthCheckPath []byte) {
	healthCheck = string(healthCheckPath)
})

var _ = SynchronizedAfterSuite(func() {
	//noop
}, func() {
	gexec.CleanupBuildArtifacts()
})
