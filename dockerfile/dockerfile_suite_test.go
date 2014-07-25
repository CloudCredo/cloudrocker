package dockerfile_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDockerfileutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dockerfile Suite")
}
