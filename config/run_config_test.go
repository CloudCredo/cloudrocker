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
			Expect(stageConfig.Command).To(Equal([]string{"/focker/fock", "stage-internal"}))
		})
	})

	Describe("Generating a RunConfig for runtime", func() {
		It("should return a valid RunConfig with the correct runtime information", func() {
			runtimeConfig := config.NewRuntimeRunConfig("fixtures/testdroplet")
			Expect(runtimeConfig.ContainerName).To(Equal("cloudfocker-runtime"))
			Expect(runtimeConfig.ImageTag).To(Equal("cloudfocker-base:latest"))
			Expect(runtimeConfig.PublishedPorts).To(Equal(map[int]int{8080: 8080}))
			Expect(len(runtimeConfig.Mounts)).To(Equal(1))
			Expect(runtimeConfig.Mounts["fixtures/testdroplet/app"]).To(Equal("/app"))
			Expect(runtimeConfig.Command).To(Equal([]string{"/bin/bash",
				"/app/cloudfocker-start.sh",
				"/app",
				"bundle", "exec", "rackup", "config.ru", "-p", "$PORT"}))
			Expect(runtimeConfig.Daemon).To(Equal(true))
			Expect(len(runtimeConfig.EnvVars)).To(Equal(3))
			Expect(runtimeConfig.EnvVars["HOME"]).To(Equal("/app"))
			Expect(runtimeConfig.EnvVars["TMPDIR"]).To(Equal("/app/tmp"))
			Expect(runtimeConfig.EnvVars["PORT"]).To(Equal("8080"))
		})
	})
})
