package config_test

import (
	"github.com/hatofmonkeys/cloudfocker/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunConfig", func() {
	Describe("Generating a RunConfig for staging", func() {
		It("should return a valid RunConfig with the correct staging information", func() {
			stageConfig := config.NewStageRunConfig("/home/testuser/testapp")
			Expect(stageConfig.ContainerName).To(Equal("cloudfocker-staging"))
			Expect(stageConfig.ImageTag).To(Equal("cloudfocker-base:latest"))
			Expect(len(stageConfig.Mounts)).To(Equal(6))
			Expect(stageConfig.Mounts["/home/testuser/testapp"]).To(Equal("/app"))
			Expect(stageConfig.Command).To(Equal([]string{"/focker/fock", "stage"}))
		})
	})
})
