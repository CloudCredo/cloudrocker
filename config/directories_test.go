package config_test

import (
	"github.com/cloudcredo/cloudfocker/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Directories", func() {
	Describe("Provide a structure for directories", func() {
		var (
			cloudFockerHomeDir string
			testDirectories    *config.Directories
		)

		BeforeEach(func() {
			cloudFockerHomeDir = "/path/to/buildpacks"
			testDirectories = config.NewDirectories(cloudFockerHomeDir)
		})

		It("should return the buildpacks directory", func() {
			Expect(testDirectories.Buildpacks()).To(Equal(cloudFockerHomeDir + "/buildpacks"))
		})

		It("should return the droplet directory", func() {
			Expect(testDirectories.Droplet()).To(Equal(cloudFockerHomeDir + "/droplet"))
		})
	})
})
