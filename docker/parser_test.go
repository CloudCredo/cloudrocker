package docker_test

import (
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"github.com/cloudcredo/cloudfocker/config"
	"github.com/cloudcredo/cloudfocker/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	Describe("Parsing a ContainerConfig for a Docker run command", func() {
		Context("with a staging config ", func() {
			It("should return a slice with all required arguments", func() {
				os.Setenv("CLOUDFOCKER_HOME", "/home/testuser/.cloudfocker")
				thisUser, _ := user.Current()
				userId := thisUser.Uid
				stageConfig := config.NewStageContainerConfig(config.NewDirectories("/home/testuser/.cloudfocker"))
				parsedRunCommand := docker.ParseRunCommand(stageConfig)
				Expect(strings.Join(parsedRunCommand, " ")).To(Equal("-u=" + userId +
					" --name=cloudfocker-staging " +
					"--volume=/home/testuser/.cloudfocker/buildpacks:/cloudfockerbuildpacks " +
					"--volume=/home/testuser/.cloudfocker/focker:/focker " +
					"--volume=/home/testuser/.cloudfocker/staging:/app " +
					"--volume=/home/testuser/.cloudfocker/tmp:/tmp " +
					"cloudfocker-base:latest " +
					"/focker/fock stage internal"))
			})
		})
		Context("with a runtime config ", func() {
			It("should return a slice with all required arguments", func() {
				os.Setenv("CLOUDFOCKER_HOME", "/home/testuser/.cloudfocker")
				thisUser, _ := user.Current()
				userId := thisUser.Uid
				testRuntimeContainerConfig := testRuntimeContainerConfig()
				parsedRunCommand := docker.ParseRunCommand(testRuntimeContainerConfig)
				Expect(strings.Join(parsedRunCommand, " ")).To(Equal("-u=" + userId +
					" --name=cloudfocker-runtime -d " +
					"--volume=/home/testuser/testapp/app:/app " +
					"--publish=8080:8080 " +
					"--env=\"HOME=/app\" " +
					"--env=\"PORT=8080\" " +
					"--env=\"TMPDIR=/app/tmp\" " +
					"cloudfocker-base:latest " +
					"/bin/bash /app/cloudfocker-start.sh /app test test test"))
			})
		})
	})
	Describe("Parsing a ContainerConfig for a Docker run command", func() {
		Context("with a runtime config ", func() {
			It("should write a valid Dockerfile", func() {
				tmpDropletDir, err := ioutil.TempDir(os.TempDir(), "parser-test-tmp-droplet")
				Expect(err).ShouldNot(HaveOccurred())
				testRuntimeContainerConfig := testRuntimeContainerConfig()
				testRuntimeContainerConfig.DropletDir = tmpDropletDir

				docker.WriteRuntimeDockerfile(testRuntimeContainerConfig)

				expected, err := ioutil.ReadFile("fixtures/build/Dockerfile")
				Expect(err).ShouldNot(HaveOccurred())
				result, err := ioutil.ReadFile(tmpDropletDir + "/Dockerfile")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result).To(Equal(expected))

				os.RemoveAll(tmpDropletDir)
			})
		})
	})
})

func testRuntimeContainerConfig() (containerConfig *config.ContainerConfig) {
	containerConfig = &config.ContainerConfig{
		ContainerName:  "cloudfocker-runtime",
		ImageTag:       "cloudfocker-base:latest",
		PublishedPorts: map[int]int{8080: 8080},
		Mounts: map[string]string{
			"/home/testuser/testapp" + "/app": "/app",
		},
		Command: append([]string{"/bin/bash", "/app/cloudfocker-start.sh", "/app"},
			[]string{"test", "test", "test"}...),
		Daemon: true,
		EnvVars: map[string]string{
			"HOME":   "/app",
			"TMPDIR": "/app/tmp",
			"PORT":   "8080",
		},
	}
	return
}
