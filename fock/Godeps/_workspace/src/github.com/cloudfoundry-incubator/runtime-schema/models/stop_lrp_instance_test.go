package models_test

import (
	. "github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StopLrpInstance", func() {
	var stopInstance StopLRPInstance

	stopInstancePayload := `{
		"process_guid":"some-process-guid",
    "instance_guid":"some-instance-guid",
    "index":1234
  }`

	BeforeEach(func() {
		stopInstance = StopLRPInstance{
			ProcessGuid:  "some-process-guid",
			InstanceGuid: "some-instance-guid",
			Index:        1234,
		}
	})
	Describe("ToJSON", func() {
		It("should JSONify", func() {
			json := stopInstance.ToJSON()
			Ω(string(json)).Should(MatchJSON(stopInstancePayload))
		})
	})

	Describe("NewStopLRPInstanceFromJSON", func() {
		It("returns a LRP with correct fields", func() {
			decodedStopInstance, err := NewStopLRPInstanceFromJSON([]byte(stopInstancePayload))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(decodedStopInstance).Should(Equal(stopInstance))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				decodedStopInstance, err := NewStopLRPInstanceFromJSON([]byte("aliens lol"))
				Ω(err).Should(HaveOccurred())

				Ω(decodedStopInstance).Should(BeZero())
			})
		})

		for field, payload := range map[string]string{
			"process_guid":  `{"instance_guid": "instance_guid", "executor_id": "executor_id"}`,
			"instance_guid": `{"process_guid": "process-guid", "executor_id": "executor_id"}`,
		} {
			json := payload
			missingField := field

			Context("when the json is missing a "+missingField, func() {
				It("returns an error indicating so", func() {
					decodedStartAuction, err := NewStopLRPInstanceFromJSON([]byte(json))
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(Equal("JSON has missing/invalid field: " + missingField))

					Ω(decodedStartAuction).Should(BeZero())
				})
			})
		}
	})

	Describe("LRPIdentifier", func() {
		It("should return a valid LRPIdentifier", func() {
			Ω(stopInstance.LRPIdentifier()).Should(Equal(LRPIdentifier{
				ProcessGuid:  "some-process-guid",
				InstanceGuid: "some-instance-guid",
				Index:        1234,
			}))
		})
	})
})
