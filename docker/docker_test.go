package docker_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"

	"github.com/cloudcredo/cloudrocker/config"
	"github.com/cloudcredo/cloudrocker/docker"

	goDockerClient "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gbytes"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gexec"
)

type FakeDockerClient struct {
	versionCalled  bool
	importImageArg goDockerClient.ImportImageOptions
	buildImageArg  goDockerClient.BuildImageOptions
	cmdImportArgs  []string
	cmdRunArgs     []string
	cmdStopArgs    []string
	cmdRmArgs      []string
	cmdKillArgs    []string
	cmdBuildArgs   []string
	cmdPsCalled    bool
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

func (f *FakeDockerClient) CmdImport(args ...string) error {
	f.cmdImportArgs = args
	return nil
}

func (f *FakeDockerClient) CmdRun(args ...string) error {
	f.cmdRunArgs = args
	return nil
}

func (f *FakeDockerClient) CmdStop(args ...string) error {
	f.cmdStopArgs = args
	return nil
}

func (f *FakeDockerClient) CmdRm(args ...string) error {
	f.cmdRmArgs = args
	return nil
}

func (f *FakeDockerClient) CmdKill(args ...string) error {
	f.cmdKillArgs = args
	return nil
}

func (f *FakeDockerClient) CmdBuild(args ...string) error {
	f.cmdBuildArgs = args
	return nil
}

func (f *FakeDockerClient) CmdPs(_ ...string) error {
	f.cmdPsCalled = true
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
				Eventually(buffer).Should(gbytes.Say("Created image."))
			})
		})
	})

	Describe("Running a configured container", func() {
		It("should tell Docker to run the container with the correct arguments", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.RunConfiguredContainer(fakeDockerClient, stdout, stdoutPipe, buffer, config.NewStageContainerConfig(config.NewDirectories("test")))
			Expect(len(fakeDockerClient.cmdRunArgs)).To(Equal(11))
			Expect(fakeDockerClient.cmdRunArgs[10]).To(Equal("internal"))
		})
	})

	Describe("Stopping the docker container", func() {
		It("should tell Docker to stop the container", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.StopContainer(fakeDockerClient, stdout, stdoutPipe, buffer, "cloudrocker-container")
			Expect(len(fakeDockerClient.cmdStopArgs)).To(Equal(1))
			Expect(fakeDockerClient.cmdStopArgs[0]).To(Equal("cloudrocker-container"))
		})
	})

	Describe("Killing the docker container", func() {
		It("should tell Docker to kill the container", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.KillContainer(fakeDockerClient, stdout, stdoutPipe, buffer, "cloudrocker-container")
			Expect(len(fakeDockerClient.cmdKillArgs)).To(Equal(1))
			Expect(fakeDockerClient.cmdKillArgs[0]).To(Equal("cloudrocker-container"))
		})
	})

	Describe("Deleting the docker container", func() {
		It("should tell Docker to delete the container", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.DeleteContainer(fakeDockerClient, stdout, stdoutPipe, buffer, "cloudrocker-container")
			Expect(len(fakeDockerClient.cmdRmArgs)).To(Equal(1))
			Expect(fakeDockerClient.cmdRmArgs[0]).To(Equal("cloudrocker-container"))
		})
	})

	Describe("Building a runtime image", func() {
		var (
			dropletDir       string
			fakeDockerClient *FakeDockerClient
			stdout           *io.PipeReader
			stdoutPipe       *io.PipeWriter
		)

		BeforeEach(func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe = io.Pipe()
			tmpDir, _ := ioutil.TempDir(os.TempDir(), "docker-runtime-image-test")
			cp("fixtures/build/droplet", tmpDir)
			dropletDir = tmpDir + "/droplet"
		})

		Context("without an image tag", func() {
			BeforeEach(func() {
				docker.BuildRuntimeImage(fakeDockerClient, stdout, stdoutPipe, buffer, config.NewRuntimeContainerConfig(dropletDir))
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
				Expect(len(fakeDockerClient.cmdBuildArgs)).To(Equal(2))
				Expect(fakeDockerClient.cmdBuildArgs[0]).To(Equal(`--tag="cloudrocker-build:latest"`))
				Expect(fakeDockerClient.cmdBuildArgs[1]).To(Equal(dropletDir))
			})
		})
		Context("with an image tag", func() {
			It("should tell Docker to build the container from the Dockerfile", func() {
				docker.BuildRuntimeImage(fakeDockerClient, stdout, stdoutPipe, buffer, config.NewRuntimeContainerConfig(dropletDir, "repository/image:tag"))
				Expect(len(fakeDockerClient.cmdBuildArgs)).To(Equal(2))
				Expect(fakeDockerClient.cmdBuildArgs[0]).To(Equal(`--tag="repository/image:tag"`))
				Expect(fakeDockerClient.cmdBuildArgs[1]).To(Equal(dropletDir))
			})
		})
	})

	Describe("Getting a cloudrocker runtime container ID", func() {
		Context("with no cloudrocker runtime container running", func() {
			It("should return empty string", func() {
				fakeDockerClient = new(FakeDockerClient)
				stdout, stdoutPipe := io.Pipe()
				containerId := make(chan string)
				go func() {
					containerId <- docker.GetContainerId(fakeDockerClient, stdout, stdoutPipe, "cloudrocker-runtime")
				}()
				io.Copy(stdoutPipe, bytes.NewBufferString("CONTAINER ID        IMAGE                COMMAND                CREATED             STATUS              PORTS                    NAMES\n"))
				Eventually(func() bool { return fakeDockerClient.cmdPsCalled }).Should(Equal(true))
				Eventually(containerId).Should(Receive(Equal("")))
			})
		})
		Context("with a cloudrocker runtime container running", func() {
			It("should return the container ID", func() {
				fakeDockerClient = new(FakeDockerClient)
				stdout, stdoutPipe := io.Pipe()
				containerId := make(chan string)
				go func() {
					containerId <- docker.GetContainerId(fakeDockerClient, stdout, stdoutPipe, "cloudrocker-runtime")
				}()
				io.Copy(stdoutPipe, bytes.NewBufferString("CONTAINER ID        IMAGE                COMMAND                CREATED             STATUS              PORTS                    NAMES\n180e16d9ef28        cloudrocker:latest   /usr/sbin/nginx -c /   13 minutes ago      Up 13 minutes       0.0.0.0:8080->8080/tcp   cloudrocker-runtime\n"))
				Eventually(func() bool { return fakeDockerClient.cmdPsCalled }).Should(Equal(true))
				Eventually(containerId).Should(Receive(Equal("180e16d9ef28")))
			})
		})
	})

	Describe("Getting a Docker client", func() {
		Context("REALDOCKER", func() {
			It("should return a usable docker client on unix", func() {
				cli := docker.GetNewClient()
				Expect(cli.Endpoint()).To(Equal("unix:///var/run/docker.sock"))
			})
		})
	})

	Describe("Container I/O plumbing", func() {
		It("Copies from a pipe to a writer without waiting for the pipe to close", func() {
			stdout, stdoutPipe := io.Pipe()

			go func() {
				docker.CopyFromPipeToPipe(buffer, stdout)
			}()

			io.Copy(stdoutPipe, bytes.NewBufferString("THIS IS A TEST STRING\n"))
			Eventually(buffer).Should(gbytes.Say(`THIS IS A TEST STRING`))

			io.Copy(stdoutPipe, bytes.NewBufferString("THIS IS ANOTHER TEST STRING\n"))
			stdoutPipe.Close()
			Eventually(buffer).Should(gbytes.Say(`THIS IS ANOTHER TEST STRING`))
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
	userId := thisUser.Uid
	return []byte(`FROM cloudrocker-raw:latest
RUN id vcap || /usr/sbin/useradd -mU -u ` + userId + ` -d /app -s /bin/bash vcap
RUN mkdir -p /app/tmp && chown -R vcap:vcap /app
`)
}
