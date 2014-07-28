package config_test

import (
	"github.com/hatofmonkeys/cloudfocker/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildConfig", func() {
	Describe("Generating a BuildConfig for staging", func() {
		It("should return a valid BuildConfig with the correct staging information", func() {
			buildConfig := config.NewStageBuildConfig()
			Expect(buildConfig.ImageTag).To(Equal("cloudfocker-base:latest"))
			Expect(buildConfig.AddFiles["fock"]).To(Equal("/"))
			Expect(buildConfig.StartCommand).To(Equal([]string{"/fock", "stage"}))
		})
	})
})
