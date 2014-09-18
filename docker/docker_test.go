package docker_test

import (
	"bytes"
	"io"

	"github.com/cloudcredo/cloudfocker/config"
	"github.com/cloudcredo/cloudfocker/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

type FakeDockerClient struct {
	cmdVersionCalled bool
	cmdImportArgs    []string
	cmdRunArgs       []string
	cmdStopArgs      []string
	cmdRmArgs        []string
	cmdKillArgs      []string
	cmdPsCalled      bool
}

func (f *FakeDockerClient) CmdVersion(_ ...string) error {
	f.cmdVersionCalled = true
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
			stdout, stdoutPipe := io.Pipe()
			docker.PrintVersion(fakeDockerClient, stdout, stdoutPipe, buffer)
			Expect(fakeDockerClient.cmdVersionCalled).To(Equal(true))
		})
	})

	Describe("Bootstrapping the Docker environment", func() {
		It("should tell Docker to import the rootfs from the supplied URL", func() {
			url := "http://test.com/test-img"
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.ImportRootfsImage(fakeDockerClient, stdout, stdoutPipe, buffer, url)
			Expect(len(fakeDockerClient.cmdImportArgs)).To(Equal(2))
			Expect(fakeDockerClient.cmdImportArgs[0]).To(Equal("http://test.com/test-img"))
			Expect(fakeDockerClient.cmdImportArgs[1]).To(Equal("cloudfocker-base"))
		})
	})

	Describe("Running a configured container", func() {
		It("should tell Docker to run the container with the correct arguments", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.RunConfiguredContainer(fakeDockerClient, stdout, stdoutPipe, buffer, config.NewStageContainerConfig(config.NewDirectories("test")))
			Expect(len(fakeDockerClient.cmdRunArgs)).To(Equal(12))
			Expect(fakeDockerClient.cmdRunArgs[11]).To(Equal("internal"))
		})
	})

	Describe("Stopping the docker container", func() {
		It("should tell Docker to stop the container", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.StopContainer(fakeDockerClient, stdout, stdoutPipe, buffer, "cloudfocker-container")
			Expect(len(fakeDockerClient.cmdStopArgs)).To(Equal(1))
			Expect(fakeDockerClient.cmdStopArgs[0]).To(Equal("cloudfocker-container"))
		})
	})

	Describe("Killing the docker container", func() {
		It("should tell Docker to kill the container", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.KillContainer(fakeDockerClient, stdout, stdoutPipe, buffer, "cloudfocker-container")
			Expect(len(fakeDockerClient.cmdKillArgs)).To(Equal(1))
			Expect(fakeDockerClient.cmdKillArgs[0]).To(Equal("cloudfocker-container"))
		})
	})

	Describe("Deleting the docker container", func() {
		It("should tell Docker to delete the container", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.DeleteContainer(fakeDockerClient, stdout, stdoutPipe, buffer, "cloudfocker-container")
			Expect(len(fakeDockerClient.cmdRmArgs)).To(Equal(1))
			Expect(fakeDockerClient.cmdRmArgs[0]).To(Equal("cloudfocker-container"))
		})
	})

	Describe("Getting a cloudfocker runtime container ID", func() {
		Context("with no cloudfocker runtime container running", func() {
			It("should return empty string", func() {
				fakeDockerClient = new(FakeDockerClient)
				stdout, stdoutPipe := io.Pipe()
				containerId := make(chan string)
				go func() {
					containerId <- docker.GetContainerId(fakeDockerClient, stdout, stdoutPipe, "cloudfocker-runtime")
				}()
				io.Copy(stdoutPipe, bytes.NewBufferString("CONTAINER ID        IMAGE                COMMAND                CREATED             STATUS              PORTS                    NAMES\n"))
				Eventually(fakeDockerClient.cmdPsCalled).Should(Equal(true))
				Eventually(containerId).Should(Receive(Equal("")))
			})
		})
		Context("with a cloudfocker runtime container running", func() {
			It("should return the container ID", func() {
				fakeDockerClient = new(FakeDockerClient)
				stdout, stdoutPipe := io.Pipe()
				containerId := make(chan string)
				go func() {
					containerId <- docker.GetContainerId(fakeDockerClient, stdout, stdoutPipe, "cloudfocker-runtime")
				}()
				io.Copy(stdoutPipe, bytes.NewBufferString("CONTAINER ID        IMAGE                COMMAND                CREATED             STATUS              PORTS                    NAMES\n180e16d9ef28        cloudfocker:latest   /usr/sbin/nginx -c /   13 minutes ago      Up 13 minutes       0.0.0.0:8080->8080/tcp   cloudfocker-runtime\n"))
				Eventually(fakeDockerClient.cmdPsCalled).Should(Equal(true))
				Eventually(containerId).Should(Receive(Equal("180e16d9ef28")))
			})
		})
	})

	Describe("Getting a Docker client", func() {
		It("should return a usable docker client on unix", func() {
			cli, stdout, stdoutpipe := docker.GetNewClient()
			docker.PrintVersion(cli, stdout, stdoutpipe, buffer)
			Eventually(buffer).Should(gbytes.Say(`Client API version: `))
		})
	})

	Describe("Container I/O plumbing", func() {
		It("Copies from a pipe to a writer", func() {
			stdout, stdoutPipe := io.Pipe()

			go func() {
				docker.CopyFromPipeToPipe(buffer, stdout)
			}()

			io.Copy(stdoutPipe, bytes.NewBufferString("THIS IS A TEST STRING\n"))
			stdoutPipe.Close()

			Eventually(buffer).Should(gbytes.Say(`THIS IS A TEST STRING`))
		})
	})
})
