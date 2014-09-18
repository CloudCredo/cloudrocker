package config_test

import (
	"github.com/cloudcredo/cloudfocker/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Directories", func() {
	var (
		cloudFockerHomeDir string
		testDirectories    *config.Directories
	)

	BeforeEach(func() {
		cloudFockerHomeDir = "/path/to"
		testDirectories = config.NewDirectories(cloudFockerHomeDir)
	})

	Describe("Provide a structure for directories", func() {
		It("should return the cloudfocker home directory", func() {
			Expect(testDirectories.Home()).To(Equal(cloudFockerHomeDir))
		})
		It("should return the buildpacks directory", func() {
			Expect(testDirectories.Buildpacks()).To(Equal(cloudFockerHomeDir + "/buildpacks"))
		})

		It("should return the droplet directory", func() {
			Expect(testDirectories.Droplet()).To(Equal(cloudFockerHomeDir + "/droplet"))
		})

		It("should return the result directory", func() {
			Expect(testDirectories.Result()).To(Equal(cloudFockerHomeDir + "/result"))
		})

		It("should return the cache directory", func() {
			Expect(testDirectories.Cache()).To(Equal(cloudFockerHomeDir + "/cache"))
		})

		It("should return the focker directory", func() {
			Expect(testDirectories.Focker()).To(Equal(cloudFockerHomeDir + "/focker"))
		})

		It("should return the staging directory", func() {
			Expect(testDirectories.Staging()).To(Equal(cloudFockerHomeDir + "/staging"))
		})
	})

	Describe("Providing the directories to be mounted in the container", func() {
		It("should return a mapping of host to container directories", func() {
			Expect(testDirectories.Mounts()).To(Equal(map[string]string{ // host dir: container dir
				"/path/to/droplet":    "/tmp/droplet",
				"/path/to/result":     "/tmp/result",
				"/path/to/buildpacks": "/tmp/cloudfockerbuildpacks",
				"/path/to/cache":      "/tmp/cache",
				"/path/to/focker":     "/focker",
			}))
		})
	})
})
