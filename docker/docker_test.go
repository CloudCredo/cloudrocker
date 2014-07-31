package docker_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/hatofmonkeys/cloudfocker/config"
	"github.com/hatofmonkeys/cloudfocker/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

type FakeDockerClient struct {
	cmdVersionCalled bool
	cmdImportArgs    []string
	cmdBuildArgs     []string
	cmdRunArgs       []string
	cmdStopArgs      []string
	cmdRmArgs        []string
	cmdKillArgs      []string
}

func (f *FakeDockerClient) CmdVersion(_ ...string) error {
	f.cmdVersionCalled = true
	return nil
}

func (f *FakeDockerClient) CmdImport(args ...string) error {
	f.cmdImportArgs = args
	return nil
}

func (f *FakeDockerClient) CmdBuild(args ...string) error {
	f.cmdBuildArgs = args
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

	Describe("Building a Docker image", func() {
		It("should ask Docker to build an image from a Dockerfile", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			location := os.TempDir() + "/testCloudFockerFile123"
			ioutil.WriteFile(location, []byte("FROM hello"), 0644)
			docker.BuildImage(fakeDockerClient, stdout, stdoutPipe, buffer, location)
			Expect(len(fakeDockerClient.cmdBuildArgs)).To(Equal(2))
			Expect(fakeDockerClient.cmdBuildArgs[0]).To(Equal("--tag=cloudfocker"))
			Expect(fakeDockerClient.cmdBuildArgs[1]).To(ContainSubstring(os.TempDir() + "/cfockerbuilder"))
			os.Remove(location)
		})
	})

	Describe("Running the docker container", func() {
		It("should tell Docker to run the container", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.RunContainer(fakeDockerClient, stdout, stdoutPipe, buffer)
			Expect(len(fakeDockerClient.cmdRunArgs)).To(Equal(4))
			Expect(fakeDockerClient.cmdRunArgs[2]).To(Equal("--name=cloudfocker-container"))
			Expect(fakeDockerClient.cmdRunArgs[3]).To(Equal("cloudfocker:latest"))
		})
	})

	Describe("Running a configured container", func() {
		It("should tell Docker to run the container with the correct arguments", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.RunConfiguredContainer(fakeDockerClient, stdout, stdoutPipe, buffer, config.NewStageRunConfig("/tmp/fakeappdir"))
			Expect(len(fakeDockerClient.cmdRunArgs)).To(Equal(11))
			Expect(fakeDockerClient.cmdRunArgs[10]).To(Equal("stage-internal"))
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

	Describe("Getting a Docker client", func() {
		It("should return a usable docker client on unix", func() {
			cli, stdout, stdoutpipe := docker.GetNewClient()
			docker.PrintVersion(cli, stdout, stdoutpipe, buffer)
			Eventually(buffer).Should(gbytes.Say(`Client API version: `))
		})
	})

	Describe("Printing to stdout", func() {
		It("should print from a pipe", func() {
			stdout, stdoutPipe := io.Pipe()
			go func() {
				docker.PrintToStdout(stdout, stdoutPipe, "stoptag", buffer)
			}()
			io.Copy(stdoutPipe, bytes.NewBufferString("THIS IS A TEST STRING\n"))
			Eventually(buffer).Should(gbytes.Say(`THIS IS A TEST STRING`))
		})
		It("should stop printing when it reaches a stoptag", func() {
			stdout, stdoutPipe := io.Pipe()
			go func() {
				docker.PrintToStdout(stdout, stdoutPipe, "stoptag", buffer)
			}()
			io.Copy(stdoutPipe, bytes.NewBufferString("THIS IS A TEST STRING\n"))
			io.Copy(stdoutPipe, bytes.NewBufferString("stoptag\n"))
			io.Copy(stdoutPipe, bytes.NewBufferString("THIS IS A NAUGHTY TEST STRING\n"))
			Eventually(buffer).Should(gbytes.Say(`THIS IS A TEST STRING`))
			Consistently(buffer).ShouldNot(gbytes.Say(`THIS IS A NAUGHTY TEST STRING`))
		})
	})
})
