package matchers_test

import (
	. "github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/onsi/gomega"
	. "github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/onsi/gomega/matchers"
)

var _ = Describe("BeTrue", func() {
	It("should handle true and false correctly", func() {
		Ω(true).Should(BeTrue())
		Ω(false).ShouldNot(BeTrue())
	})

	It("should only support booleans", func() {
		success, err := (&BeTrueMatcher{}).Match("foo")
		Ω(success).Should(BeFalse())
		Ω(err).Should(HaveOccurred())
	})
})
