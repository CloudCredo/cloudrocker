package models_test

import (
	. "github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LRPIdentifier", func() {
	Describe("generating an LRPIdentifier from an OpaqueID representation", func() {
		It("should succeed when the representation is well formed", func() {
			identifier, err := LRPIdentifierFromOpaqueID("process-guid.17.instance-guid")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(identifier).Should(Equal(LRPIdentifier{
				ProcessGuid:  "process-guid",
				Index:        17,
				InstanceGuid: "instance-guid",
			}))
		})

		It("should error when there are not three subcomponents", func() {
			identifier, err := LRPIdentifierFromOpaqueID("process-guid.17.instance-guid.foo")
			Ω(err).Should(HaveOccurred())
			Ω(identifier).Should(BeZero())

			identifier, err = LRPIdentifierFromOpaqueID("process-guid.17")
			Ω(err).Should(HaveOccurred())
			Ω(identifier).Should(BeZero())
		})

		It("should error when the index is not an integer", func() {
			identifier, err := LRPIdentifierFromOpaqueID("process-guid.seventeen.instance-guid")
			Ω(err).Should(HaveOccurred())
			Ω(identifier).Should(BeZero())
		})
	})
})
