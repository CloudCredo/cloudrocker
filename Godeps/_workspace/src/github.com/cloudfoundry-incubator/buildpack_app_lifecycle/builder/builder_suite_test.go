package main_test

import (
	"testing"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gexec"
)

var builderPath string

func TestBuildpackLifecycleBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildpack-Lifecycle-Builder Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	builder, err := gexec.Build("github.com/cloudfoundry-incubator/buildpack_app_lifecycle/builder")
	Ω(err).ShouldNot(HaveOccurred())
	return []byte(builder)
}, func(builder []byte) {
	builderPath = string(builder)
})

var _ = SynchronizedAfterSuite(func() {
	//noop
}, func() {
	gexec.CleanupBuildArtifacts()
})
