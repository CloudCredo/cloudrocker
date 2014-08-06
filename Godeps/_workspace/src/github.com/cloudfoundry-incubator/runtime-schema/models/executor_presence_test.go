package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("ExecutorPresence", func() {
	var executorPresence ExecutorPresence

	const payload = `{
    "executor_id":"some-id",
    "stack": "some-stack"
  }`

	BeforeEach(func() {
		executorPresence = ExecutorPresence{
			ExecutorID: "some-id",
			Stack:      "some-stack",
		}
	})

	Describe("ToJSON", func() {
		It("should JSONify", func() {
			json := executorPresence.ToJSON()
			Ω(string(json)).Should(MatchJSON(payload))
		})
	})

	Describe("NewTaskFromJSON", func() {
		It("returns a Task with correct fields", func() {
			decodedExecutorPresence, err := NewExecutorPresenceFromJSON([]byte(payload))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(decodedExecutorPresence).Should(Equal(executorPresence))
		})

		Context("with an invalid payload", func() {
			It("returns the error", func() {
				decodedExecutorPresence, err := NewExecutorPresenceFromJSON([]byte("aliens lol"))
				Ω(err).Should(HaveOccurred())

				Ω(decodedExecutorPresence).Should(BeZero())
			})
		})
	})
})
