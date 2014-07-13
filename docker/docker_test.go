package docker_test

import (
	"bytes"
	"io"
	"os"

	"github.com/hatofmonkeys/cloudfocker/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

type FakeDockerClient struct {
	cmdVersionCalled bool
}

func (f *FakeDockerClient) CmdVersion(_ ...string) error {
	f.cmdVersionCalled = true
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

	Describe("Getting a rootfs URL", func() {
		Context("without a rootfs env var set", func() {
			It("should return the default URL", func() {
				os.Setenv("FOCKER_ROOTFS_URL", "")
				Expect(docker.GetRootfsUrl()).To(Equal("https://s3.amazonaws.com/blob.cfblob.com/fee97b71-17d7-4fab-a5b0-69d4112521e6"))
			})
		})
		Context("with a rootfs env var set", func() {
			It("should return the specified URL", func() {
				os.Setenv("FOCKER_ROOTFS_URL", "dave")
				Expect(docker.GetRootfsUrl()).To(Equal("dave"))
			})
		})
	})

	Describe("Displaying the Docker version", func() {
		It("should tell Docker to output its version", func() {
			fakeDockerClient = new(FakeDockerClient)
			stdout, stdoutPipe := io.Pipe()
			docker.PrintVersion(fakeDockerClient, stdout, stdoutPipe, buffer)
			Expect(fakeDockerClient.cmdVersionCalled).To(Equal(true))
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
