package buildpack_app_lifecycle_test

import (
	"testing"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
)

func TestBuildpackLifecycle(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildpack-Lifecycle Suite")
}
