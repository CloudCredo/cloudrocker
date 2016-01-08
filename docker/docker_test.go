package docker_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"time"

	"github.com/cloudcredo/cloudrocker/config"
	"github.com/cloudcredo/cloudrocker/docker"

	goDockerClient "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gbytes"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gexec"
)

type FakeDockerClient struct {
	versionCalled                      bool
	importImageArg                     goDockerClient.ImportImageOptions
	buildImageArg                      goDockerClient.BuildImageOptions
	listContainersArg                  goDockerClient.ListContainersOptions
	removeContainerArg                 goDockerClient.RemoveContainerOptions
	stopContainerArgID                 string
	stopContainerArgTimeout            uint
	createContainerArg                 goDockerClient.CreateContainerOptions
	startContainerArgID                string
	startContainerArgHostConfig        *goDockerClient.HostConfig
	attachToContainerNonBlockingCalled bool
	attachToContainerNonBlockingArg    goDockerClient.AttachToContainerOptions
	addEventListenerCalled             bool
}

func (fake *FakeDockerClient) Version() (*goDockerClient.Env, error) {
	fake.versionCalled = true
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

func (fake *FakeDockerClient) ImportImage(options goDockerClient.ImportImageOptions) error {
	fake.importImageArg = options
	return nil
}

func (fake *FakeDockerClient) BuildImage(options goDockerClient.BuildImageOptions) error {
	fake.buildImageArg = options
	return nil
}

func (fake *FakeDockerClient) ListContainers(options goDockerClient.ListContainersOptions) ([]goDockerClient.APIContainers, error) {
	fake.listContainersArg = options
	containers := []goDockerClient.APIContainers{
		{
			ID:    "e8096241370a",
			Names: []string{"/cloudrocker-runtime"},
		},
	}
	return containers, nil
}

func (fake *FakeDockerClient) RemoveContainer(options goDockerClient.RemoveContainerOptions) error {
	fake.removeContainerArg = options
	return nil
}

func (fake *FakeDockerClient) StopContainer(id string, timeout uint) error {
	fake.stopContainerArgID = id
	fake.stopContainerArgTimeout = timeout
	return nil
}

func (fake *FakeDockerClient) CreateContainer(options goDockerClient.CreateContainerOptions) (*goDockerClient.Container, error) {
	fake.createContainerArg = options
	var container = goDockerClient.Container{
		ID: "5716e9326cd9",
	}
	return &container, nil
}

func (fake *FakeDockerClient) StartContainer(id string, hostConfig *goDockerClient.HostConfig) error {
	fake.startContainerArgID = id
	fake.startContainerArgHostConfig = hostConfig
	return nil
}

func (fake *FakeDockerClient) AttachToContainerNonBlocking(attachToContainerOptions goDockerClient.AttachToContainerOptions) (goDockerClient.CloseWaiter, error) {
	fake.attachToContainerNonBlockingCalled = true
	fake.attachToContainerNonBlockingArg = attachToContainerOptions
	return nil, nil
}

func (fake *FakeDockerClient) AddEventListener(listener chan<- *goDockerClient.APIEvents) error {
	fake.addEventListenerCalled = true
	go func() {
		time.Sleep(time.Nanosecond * 1000)
		listener <- &goDockerClient.APIEvents{Status: "die"}
	}()
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
				cli := docker.GetNewClient()
				Expect(cli.Endpoint()).To(Equal("unix:///var/run/docker.sock"))
			})
		})
	})

	Describe("Displaying the Docker version", func() {
		It("should tell Docker to output its version", func() {
			fakeDockerClient = new(FakeDockerClient)
			docker.PrintVersion(fakeDockerClient, buffer)
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
			docker.ImportRootfsImage(fakeDockerClient, buffer, url)
			Expect(fakeDockerClient.importImageArg.Source).To(Equal("http://test.com/test-img"))
			Expect(fakeDockerClient.importImageArg.Repository).To(Equal("cloudrocker-raw"))
			Expect(fakeDockerClient.importImageArg.OutputStream).To(Equal(buffer))
		})

		Describe("telling Docker to build a base image from the raw image with the correct config for rocker use", func() {
			var (
				buildDir string
			)
			BeforeEach(func() {
				buildDir, _ = ioutil.TempDir(os.TempDir(), "docker-configure-base")
				fakeDockerClient = new(FakeDockerClient)
				docker.BuildBaseImage(fakeDockerClient, buffer, config.NewBaseContainerConfig(buildDir))
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
				Eventually(buffer).Should(gbytes.Say("Created base image."))
			})
		})
	})

	Describe("Getting a cloudrocker container ID", func() {
		Context("when no cloudrocker container is found", func() {
			It("should return empty string", func() {
				fakeDockerClient = new(FakeDockerClient)
				containerID := docker.GetContainerID(fakeDockerClient, "another-container")
				Expect(fakeDockerClient.listContainersArg.All).To(Equal(true))
				Expect(containerID).To(Equal(""))
			})
		})

		Context("when a cloudrocker container exists", func() {
			It("should return the container ID", func() {
				fakeDockerClient = new(FakeDockerClient)
				containerID := docker.GetContainerID(fakeDockerClient, "cloudrocker-runtime")
				Expect(fakeDockerClient.listContainersArg.All).To(Equal(true))
				Expect(containerID).To(Equal("e8096241370a"))
			})
		})
	})

	Describe("Deleting the docker container", func() {
		It("should tell Docker to delete the container", func() {
			fakeDockerClient = new(FakeDockerClient)
			docker.DeleteContainer(fakeDockerClient, buffer, "cloudrocker-runtime")
			Expect(fakeDockerClient.removeContainerArg.Force).To(Equal(true))
			Expect(fakeDockerClient.removeContainerArg.ID).To(Equal("e8096241370a"))
		})
	})

	Describe("Stopping the docker container", func() {
		It("should tell Docker to stop the container", func() {
			fakeDockerClient = new(FakeDockerClient)
			docker.StopContainer(fakeDockerClient, buffer, "cloudrocker-runtime")
			Expect(fakeDockerClient.stopContainerArgID).To(Equal("e8096241370a"))
			var timeout uint = 10
			Expect(fakeDockerClient.stopContainerArgTimeout).To(Equal(timeout))
		})
	})

	Describe("Building a runtime image", func() {
		var (
			dropletDir       string
			fakeDockerClient *FakeDockerClient
		)

		BeforeEach(func() {
			fakeDockerClient = new(FakeDockerClient)
			tmpDir, _ := ioutil.TempDir(os.TempDir(), "docker-runtime-image-test")
			cp("fixtures/build/droplet", tmpDir)
			dropletDir = tmpDir + "/droplet"
		})

		Context("without an image tag", func() {
			BeforeEach(func() {
				docker.BuildRuntimeImage(fakeDockerClient, buffer, config.NewRuntimeContainerConfig(dropletDir))
			})

			It("should create a tarred version of the droplet mount, for extraction in the container, so as to not have AUFS permissions issues in https://github.com/docker/docker/issues/783", func() {
				dropletDirFile, err := os.Open(dropletDir)
				Expect(err).ShouldNot(HaveOccurred())
				dropletDirContents, err := dropletDirFile.Readdirnames(0)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(dropletDirContents, err).Should(ContainElement("droplet.tgz"))
			})

			It("should write a valid Dockerfile to the filesystem", func() {
				result, err := ioutil.ReadFile(dropletDir + "/Dockerfile")
				Expect(err).ShouldNot(HaveOccurred())
				expected, err := ioutil.ReadFile("fixtures/build/Dockerfile")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result).To(Equal(expected))
			})

			It("should tell Docker to build the container from the Dockerfile", func() {
				Expect(fakeDockerClient.buildImageArg.Name).To(Equal("cloudrocker-build:latest"))
				Expect(fakeDockerClient.buildImageArg.ContextDir).To(Equal(dropletDir))
				Expect(fakeDockerClient.buildImageArg.Dockerfile).To(Equal("/Dockerfile"))
				Expect(fakeDockerClient.buildImageArg.OutputStream).To(Equal(buffer))
				Eventually(buffer).Should(gbytes.Say("Created runtime image."))
			})
		})

		Context("with an image tag", func() {
			It("should tell Docker to build the container from the Dockerfile", func() {
				docker.BuildRuntimeImage(fakeDockerClient, buffer, config.NewRuntimeContainerConfig(dropletDir, "repository/image:tag"))

				Expect(fakeDockerClient.buildImageArg.Name).To(Equal("repository/image:tag"))
				Expect(fakeDockerClient.buildImageArg.ContextDir).To(Equal(dropletDir))
				Expect(fakeDockerClient.buildImageArg.Dockerfile).To(Equal("/Dockerfile"))
				Expect(fakeDockerClient.buildImageArg.OutputStream).To(Equal(buffer))
				Eventually(buffer).Should(gbytes.Say("Created runtime image."))
			})
		})
	})

	Describe("Running a staging container", func() {
		It("should tell Docker to run the container with the correct arguments", func() {
			thisUser, _ := user.Current()
			userID := thisUser.Uid
			fakeDockerClient = new(FakeDockerClient)

			docker.RunStagingContainer(fakeDockerClient, buffer, config.NewStageContainerConfig(config.NewDirectories("/test")))

			Expect(fakeDockerClient.createContainerArg.Name).To(Equal("cloudrocker-staging"))
			Expect(fakeDockerClient.createContainerArg.Config.User).To(Equal(userID))
			Expect(fakeDockerClient.createContainerArg.Config.Env).To(Equal([]string{"CF_STACK=cflinuxfs2"}))
			Expect(fakeDockerClient.createContainerArg.Config.Image).To(Equal("cloudrocker-base:latest"))
			Expect(fakeDockerClient.createContainerArg.Config.Cmd).To(Equal([]string{"/rocker/rock", "stage", "internal"}))
			var mounts = []goDockerClient.Mount{
				goDockerClient.Mount{
					Source:      "/test/buildpacks",
					Destination: "/cloudrockerbuildpacks",
					RW:          true,
				},
				goDockerClient.Mount{
					Source:      "/test/rocker",
					Destination: "/rocker",
					RW:          true,
				},
				goDockerClient.Mount{
					Source:      "/test/staging",
					Destination: "/tmp/app",
					RW:          true,
				},
				goDockerClient.Mount{
					Source:      "/test/tmp",
					Destination: "/tmp",
					RW:          true,
				},
			}
			Expect(fakeDockerClient.createContainerArg.Config.Mounts).To(Equal(mounts))
			Expect(fakeDockerClient.createContainerArg.Config.AttachStdout).To(Equal(true))
			Expect(fakeDockerClient.createContainerArg.Config.AttachStderr).To(Equal(true))
			var binds = []string{
				"/test/buildpacks:/cloudrockerbuildpacks",
				"/test/rocker:/rocker",
				"/test/staging:/tmp/app",
				"/test/tmp:/tmp",
			}
			Expect(fakeDockerClient.createContainerArg.HostConfig.Binds).To(Equal(binds))
			Expect(fakeDockerClient.createContainerArg.HostConfig.NetworkMode).To(Equal("bridge"))

			Expect(fakeDockerClient.attachToContainerNonBlockingArg).To(Equal(goDockerClient.AttachToContainerOptions{
				Container:    "5716e9326cd9",
				OutputStream: buffer,
				Stdout:       true,
				Stderr:       true,
				Stream:       true,
			}))
			Expect(fakeDockerClient.addEventListenerCalled).To(Equal(true))
			Expect(fakeDockerClient.startContainerArgID).To(Equal("5716e9326cd9"))
			var noHostConfig *goDockerClient.HostConfig
			Expect(fakeDockerClient.startContainerArgHostConfig).To(Equal(noHostConfig))
		})
	})

	Describe("Running a runtime container", func() {
		It("should tell Docker to run the container with the correct arguments", func() {
			thisUser, _ := user.Current()
			userID := thisUser.Uid
			fakeDockerClient = new(FakeDockerClient)

			docker.RunRuntimeContainer(fakeDockerClient, buffer, testRuntimeContainerConfig())

			Expect(fakeDockerClient.createContainerArg.Name).To(Equal("cloudrocker-runtime"))
			Expect(fakeDockerClient.createContainerArg.Config.User).To(Equal(userID))
			Expect(fakeDockerClient.createContainerArg.Config.Env).To(Equal([]string{
				"HOME=/app",
				"PORT=8080",
				"TMPDIR=/app/tmp",
			}))
			Expect(fakeDockerClient.createContainerArg.Config.Image).To(Equal("cloudrocker-base:latest"))
			Expect(fakeDockerClient.createContainerArg.Config.Cmd).To(Equal([]string{
				"/bin/bash",
				"/app/cloudrocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh",
				"/app",
				"the",
				"start",
				"command",
				"\"quoted",
				"string",
				"with",
				"spaces\"",
			}))
			var mounts = []goDockerClient.Mount{
				goDockerClient.Mount{
					Source:      "/home/testuser/testapp/app",
					Destination: "/app",
					RW:          true,
				},
			}
			Expect(fakeDockerClient.createContainerArg.Config.Mounts).To(Equal(mounts))
			Expect(fakeDockerClient.createContainerArg.Config.AttachStdout).To(Equal(false))
			Expect(fakeDockerClient.createContainerArg.Config.AttachStderr).To(Equal(false))
			Expect(fakeDockerClient.createContainerArg.Config.ExposedPorts).To(Equal(map[goDockerClient.Port]struct{}{
				"8080/tcp": struct{}{},
			}))
			Expect(fakeDockerClient.createContainerArg.HostConfig.Binds).To(Equal([]string{
				"/home/testuser/testapp/app:/app",
			}))
			var portBindings = map[goDockerClient.Port][]goDockerClient.PortBinding{
				"8080/tcp": []goDockerClient.PortBinding{
					{
						HostPort: "8080",
					},
				},
			}
			Expect(fakeDockerClient.createContainerArg.HostConfig.PortBindings).To(Equal(portBindings))
			Expect(fakeDockerClient.createContainerArg.HostConfig.NetworkMode).To(Equal("bridge"))

			Expect(fakeDockerClient.attachToContainerNonBlockingCalled).To(Equal(false))
			Expect(fakeDockerClient.addEventListenerCalled).To(Equal(false))
			Expect(fakeDockerClient.startContainerArgID).To(Equal("5716e9326cd9"))
			var noHostConfig *goDockerClient.HostConfig
			Expect(fakeDockerClient.startContainerArgHostConfig).To(Equal(noHostConfig))
		})
	})
})

func cp(src string, dst string) {
	session, err := gexec.Start(
		exec.Command("cp", "-a", src, dst),
		GinkgoWriter,
		GinkgoWriter,
	)
	Ω(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
}

func buildBaseImageDockerfile() []byte {
	thisUser, _ := user.Current()
	userID := thisUser.Uid
	return []byte(`FROM cloudrocker-raw:latest
RUN id vcap || /usr/sbin/useradd -mU -u ` + userID + ` -d /app -s /bin/bash vcap
RUN mkdir -p /app/tmp && chown -R vcap:vcap /app
`)
}
