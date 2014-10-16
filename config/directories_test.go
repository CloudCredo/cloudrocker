package config_test

import (
	"os"

	"github.com/cloudcredo/cloudfocker/config"

	. "github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/onsi/gomega"
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

		It("should return the container's buildpacks directory", func() {
			Expect(testDirectories.ContainerBuildpacks()).To(Equal("/cloudfockerbuildpacks"))
		})

		It("should return the droplet directory", func() {
			Expect(testDirectories.Droplet()).To(Equal(cloudFockerHomeDir + "/tmp/droplet"))
		})

		It("should return the result directory", func() {
			Expect(testDirectories.Result()).To(Equal(cloudFockerHomeDir + "/tmp/result"))
		})

		It("should return the cache directory", func() {
			Expect(testDirectories.Cache()).To(Equal(cloudFockerHomeDir + "/tmp/cache"))
		})

		It("should return the focker directory", func() {
			Expect(testDirectories.Focker()).To(Equal(cloudFockerHomeDir + "/focker"))
		})

		It("should return the staging directory", func() {
			Expect(testDirectories.Staging()).To(Equal(cloudFockerHomeDir + "/staging"))
		})

		It("should return the host cloudfocker tmp directory", func() {
			Expect(testDirectories.Tmp()).To(Equal(cloudFockerHomeDir + "/tmp"))
		})

		It("should return the application directory", func() {
			pwd, _ := os.Getwd()
			Expect(testDirectories.App()).To(Equal(pwd))
		})
	})

	Describe("Providing the directories to be mounted in the container", func() {
		It("should return a mapping of host to container directories", func() {
			Expect(testDirectories.Mounts()).To(Equal(map[string]string{ // host dir: container dir
				"/path/to/tmp":        "/tmp",
				"/path/to/focker":     "/focker",
				"/path/to/buildpacks": "/cloudfockerbuildpacks",
				"/path/to/staging":    "/app",
			}))
		})
	})

	Describe("Providing the directories to be created before staging", func() {
		It("should return a set of directories to be created", func() {
			Expect(testDirectories.HostDirectories()).To(ConsistOf(
				"/path/to",
				"/path/to/buildpacks",
				"/path/to/tmp/droplet",
				"/path/to/tmp/result",
				"/path/to/tmp/cache",
				"/path/to/focker",
				"/path/to/staging",
				"/path/to/tmp",
			))
		})
	})

	Describe("Providing the directories to be cleaned before staging", func() {
		It("should return a set of directories to be cleaned", func() {
			Expect(testDirectories.HostDirectoriesToClean()).To(ConsistOf(
				"/path/to/tmp/droplet",
				"/path/to/tmp/result",
				"/path/to/staging",
			))
		})
	})
})
