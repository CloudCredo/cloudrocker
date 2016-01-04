package godocker_test

import (
	"io/ioutil"
	"os"
	"os/user"

	"github.com/cloudcredo/cloudrocker/config"
	"github.com/cloudcredo/cloudrocker/godocker"

	goDockerClient "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gbytes"
)

type FakeDockerClient struct {
	versionCalled  bool
	importImageArg goDockerClient.ImportImageOptions
	buildImageArg  goDockerClient.BuildImageOptions
}

func (f *FakeDockerClient) Version() (*goDockerClient.Env, error) {
	f.versionCalled = true
	versionList := new(goDockerClient.Env)
	versionList.Set("Os", "linux")
	versionList.Set("Arch", "amd64")
	versionList.Set("GitCommit", "a8a31ef")
	versionList.Set("GoVersion", "go1.4.1")
	versionList.Set("KernelVersion", "3.13.0-24-generic")
	versionList.Set("Version", "1.5.0")
	versionList.Set("ApiVersion", "1.17")
	return versionList, nil
}

func (f *FakeDockerClient) ImportImage(options goDockerClient.ImportImageOptions) error {
	f.importImageArg = options
	return nil
}

func (f *FakeDockerClient) BuildImage(options goDockerClient.BuildImageOptions) error {
	f.buildImageArg = options
	return nil
}

var _ = Describe("Docker", func() {
	var (
		fakeDockerClient *FakeDockerClient
		buffer           *gbytes.Buffer
	)

	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
	})

	Describe("Getting a Docker client", func() {
		Context("REALDOCKER", func() {
			It("should return a usable docker client on unix", func() {
				cli := godocker.GetNewClient()
				Expect(cli.Endpoint()).To(Equal("unix:///var/run/docker.sock"))
			})
		})
	})

	Describe("Displaying the Docker version", func() {
		It("should tell Docker to output its version", func() {
			fakeDockerClient = new(FakeDockerClient)
			godocker.PrintVersion(fakeDockerClient, buffer)
			Expect(fakeDockerClient.versionCalled).To(Equal(true))
			Eventually(buffer).Should(gbytes.Say("Client OS/Arch: linux/amd64"))
			Eventually(buffer).Should(gbytes.Say("Server version: 1.5.0"))
			Eventually(buffer).Should(gbytes.Say("Server API version: 1.17"))
			Eventually(buffer).Should(gbytes.Say("Server Go version: go1.4.1"))
			Eventually(buffer).Should(gbytes.Say("Server Git commit: a8a31ef"))
		})
	})

	Describe("Bootstrapping the Docker environment", func() {
		It("should tell Docker to import the rootfs from the supplied URL", func() {
			url := "http://test.com/test-img"
			fakeDockerClient = new(FakeDockerClient)
			godocker.ImportRootfsImage(fakeDockerClient, buffer, url)
			Expect(fakeDockerClient.importImageArg.Source).To(Equal("http://test.com/test-img"))
			Expect(fakeDockerClient.importImageArg.Repository).To(Equal("cloudrocker-raw"))
			Expect(fakeDockerClient.importImageArg.Tag).To(Equal("cloudrocker-base:latest"))
			Expect(fakeDockerClient.importImageArg.OutputStream).To(Equal(buffer))
		})

		Describe("telling Docker to build a base image from the raw image with the correct config for rocker use", func() {
			var (
				buildDir string
			)
			BeforeEach(func() {
				buildDir, _ = ioutil.TempDir(os.TempDir(), "docker-configure-base")
				fakeDockerClient = new(FakeDockerClient)
				godocker.BuildBaseImage(fakeDockerClient, buffer, config.NewBaseContainerConfig(buildDir))
			})
			AfterEach(func() {
				os.RemoveAll(buildDir)
			})

			It("should write a valid and correct Dockerfile to the filesystem", func() {
				result, err := ioutil.ReadFile(buildDir + "/Dockerfile")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result).To(Equal(buildBaseImageDockerfile()))
			})

			It("should tell Docker to build the configured rootfs image from the Dockerfile", func() {
				Expect(fakeDockerClient.buildImageArg.Name).To(Equal("cloudrocker-base:latest"))
				Expect(fakeDockerClient.buildImageArg.ContextDir).To(Equal(buildDir))
				Expect(fakeDockerClient.buildImageArg.Dockerfile).To(Equal("/Dockerfile"))
				Expect(fakeDockerClient.buildImageArg.OutputStream).To(Equal(buffer))
				Eventually(buffer).Should(gbytes.Say("Created image."))
			})
		})
	})
})

func buildBaseImageDockerfile() []byte {
	thisUser, _ := user.Current()
	userID := thisUser.Uid
	return []byte(`FROM cloudrocker-raw:latest
RUN id vcap || /usr/sbin/useradd -mU -u ` + userID + ` -d /app -s /bin/bash vcap
RUN mkdir -p /app/tmp && chown -R vcap:vcap /app
`)
}
