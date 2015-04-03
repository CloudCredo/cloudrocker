package focker_test

import (
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"

	"testing"
)

func TestFocker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Focker Suite")
}
