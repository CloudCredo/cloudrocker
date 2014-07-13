package focker_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFocker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Focker Suite")
}
