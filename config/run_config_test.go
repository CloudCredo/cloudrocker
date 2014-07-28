package config_test

import (
	"github.com/hatofmonkeys/cloudfocker/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunConfig", func() {
	Describe("Generating a RunConfig for staging", func() {
		It("should return a valid RunConfig with the correct staging information", func() {
			buildConfig := config.NewStageRunConfig("/home/testuser/testapp")
			Expect(buildConfig.ContainerName).To(Equal("cloudfocker-staging"))
			Expect(buildConfig.ImageTag).To(Equal("cloudfocker-base:latest"))
			Expect(len(buildConfig.Mounts)).To(Equal(6))
			Expect(buildConfig.Mounts["/home/testuser/testapp"]).To(Equal("/app"))
			Expect(buildConfig.RunCommand).To(Equal([]string{"/focker/fock", "stage"}))
		})
	})
})
