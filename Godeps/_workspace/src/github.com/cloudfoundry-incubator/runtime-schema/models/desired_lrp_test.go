package models_test

import (
	. "github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRP", func() {
	var lrp DesiredLRP

	lrpPayload := `{
    "process_guid":"some-guid",
    "instances":5,
    "stack":"some-stack",
    "memory_mb":1024,
    "disk_mb":512,
    "file_descriptors":17,
    "source":"http://example.com/source",
    "start_command":"echo",
    "environment": [{"name": "FOO", "value": "BAR"}],
    "routes":["route-1","route-2"],
    "log_guid":"some-log-guid"
  }`

	BeforeEach(func() {
		lrp = DesiredLRP{
			ProcessGuid:     "some-guid",
			Instances:       5,
			Stack:           "some-stack",
			MemoryMB:        1024,
			DiskMB:          512,
			FileDescriptors: 17,
			Source:          "http://example.com/source",
			StartCommand:    "echo",
			Environment:     []EnvironmentVariable{{Name: "FOO", Value: "BAR"}},
			Routes:          []string{"route-1", "route-2"},
			LogGuid:         "some-log-guid",
		}
	})

	Describe("ToJSON", func() {
		It("should JSONify", func() {
			json := lrp.ToJSON()
			Ω(string(json)).Should(MatchJSON(lrpPayload))
		})
	})

	Describe("NewDesiredLRPFromJSON", func() {
		It("returns a LRP with correct fields", func() {
			decodedStartAuction, err := NewDesiredLRPFromJSON([]byte(lrpPayload))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(decodedStartAuction).Should(Equal(lrp))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				decodedStartAuction, err := NewDesiredLRPFromJSON([]byte("aliens lol"))
				Ω(err).Should(HaveOccurred())

				Ω(decodedStartAuction).Should(BeZero())
			})
		})

		for field, payload := range map[string]string{
			"process_guid": `{"source": "http://example.com", "stack": "some-stack"}`,
			"source":       `{"process_guid": "process_guid", "stack": "some-stack"}`,
			"stack":        `{"process_guid": "process_guid", "source": "http://example.com"}`,
		} {
			json := payload
			missingField := field

			Context("when the json is missing a "+missingField, func() {
				It("returns an error indicating so", func() {
					decodedStartAuction, err := NewDesiredLRPFromJSON([]byte(json))
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(Equal("JSON has missing/invalid field: " + missingField))

					Ω(decodedStartAuction).Should(BeZero())
				})
			})
		}
	})
})
