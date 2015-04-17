package config_test

import (
	"github.com/cloudcredo/cloudrocker/config"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("ContainerConfig", func() {
	Describe("Generating a ContainerConfig for staging", func() {
		It("should return a valid ContainerConfig with the correct staging information", func() {
			stageConfig := config.NewStageContainerConfig(config.NewDirectories("TEST_CLOUDROCKERHOME"))
			Expect(stageConfig.ContainerName).To(Equal("cloudrocker-staging"))
			Expect(stageConfig.Mounts["TEST_CLOUDROCKERHOME/staging"]).To(Equal("/tmp/app"))
			Expect(stageConfig.Mounts["TEST_CLOUDROCKERHOME/tmp"]).To(Equal("/tmp"))
			Expect(stageConfig.Mounts["TEST_CLOUDROCKERHOME/buildpacks"]).To(Equal("/cloudrockerbuildpacks"))
			Expect(stageConfig.Mounts["TEST_CLOUDROCKERHOME/rocker"]).To(Equal("/rocker"))
			Expect(stageConfig.ImageTag).To(Equal("cloudrocker-base:latest"))
			Expect(stageConfig.Command).To(Equal([]string{"/rocker/rock", "stage", "internal"}))
		})
	})

	Describe("Generating a ContainerConfig for runtime", func() {
		Context("with a valid staging_info.yml", func() {
			It("should return a valid ContainerConfig with the correct runtime information", func() {
				runtimeConfig := config.NewRuntimeContainerConfig("fixtures/testdroplet")
				Expect(runtimeConfig.ContainerName).To(Equal("cloudrocker-runtime"))
				Expect(runtimeConfig.Daemon).To(Equal(true))
				Expect(len(runtimeConfig.Mounts)).To(Equal(1))
				Expect(runtimeConfig.Mounts["fixtures/testdroplet/app"]).To(Equal("/app"))
				Expect(runtimeConfig.PublishedPorts).To(Equal(map[int]int{8080: 8080}))
				Expect(len(runtimeConfig.EnvVars)).To(Equal(5))
				Expect(runtimeConfig.EnvVars["HOME"]).To(Equal("/app"))
				Expect(runtimeConfig.EnvVars["PORT"]).To(Equal("8080"))
				Expect(runtimeConfig.EnvVars["TMPDIR"]).To(Equal("/app/tmp"))
				Expect(runtimeConfig.EnvVars["VCAP_SERVICES"]).To(Equal("{ \"elephantsql\": [ { \"name\": \"elephantsql-c6c60\", \"label\": \"elephantsql\", \"tags\": [ \"postgres\", \"postgresql\", \"relational\" ], \"plan\": \"turtle\", \"credentials\": { \"uri\": \"postgres://seilbmbd:PHxTPJSbkcDakfK4cYwXHiIX9Q8p5Bxn@babar.elephantsql.com:5432/seilbmbd\" } } ], \"sendgrid\": [ { \"name\": \"mysendgrid\", \"label\": \"sendgrid\", \"tags\": [ \"smtp\" ], \"plan\": \"free\", \"credentials\": { \"hostname\": \"smtp.sendgrid.net\", \"username\": \"QvsXMbJ3rK\", \"password\": \"HCHMOYluTv\" } } ] }"))
				Expect(runtimeConfig.EnvVars["DATABASE_URL"]).To(Equal("postgres://seilbmbd:PHxTPJSbkcDakfK4cYwXHiIX9Q8p5Bxn@babar.elephantsql.com:5432/seilbmbd"))
				Expect(runtimeConfig.ImageTag).To(Equal("cloudrocker-base:latest"))
				Expect(runtimeConfig.Command).To(Equal([]string{"/bin/bash",
					"/app/cloudrocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh",
					"/app",
					"bundle exec rackup config.ru -p $PORT"}))
				Expect(runtimeConfig.DropletDir).To(Equal("fixtures/testdroplet"))
			})
		})
		Context("with no staging_info.yml, but a valid Procfile", func() {
			It("should return a valid ContainerConfig with the correct runtime information", func() {
				runtimeConfig := config.NewRuntimeContainerConfig("fixtures/procfiletestdroplet")
				Expect(runtimeConfig.ContainerName).To(Equal("cloudrocker-runtime"))
				Expect(runtimeConfig.Daemon).To(Equal(true))
				Expect(len(runtimeConfig.Mounts)).To(Equal(1))
				Expect(runtimeConfig.Mounts["fixtures/procfiletestdroplet/app"]).To(Equal("/app"))
				Expect(runtimeConfig.PublishedPorts).To(Equal(map[int]int{8080: 8080}))
				Expect(len(runtimeConfig.EnvVars)).To(Equal(5))
				Expect(runtimeConfig.EnvVars["HOME"]).To(Equal("/app"))
				Expect(runtimeConfig.EnvVars["TMPDIR"]).To(Equal("/app/tmp"))
				Expect(runtimeConfig.EnvVars["PORT"]).To(Equal("8080"))
				Expect(runtimeConfig.EnvVars["VCAP_SERVICES"]).To(Equal(""))
				Expect(runtimeConfig.EnvVars["DATABASE_URL"]).To(Equal(""))
				Expect(runtimeConfig.ImageTag).To(Equal("cloudrocker-base:latest"))
				Expect(runtimeConfig.Command).To(Equal([]string{"/bin/bash",
					"/app/cloudrocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh",
					"/app",
					"server"}))
			})
		})
	})
})
