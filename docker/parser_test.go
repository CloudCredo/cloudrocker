package docker_test

import (
	"os"
	"os/user"
	"strings"

	"github.com/cloudcredo/cloudfocker/config"
	"github.com/cloudcredo/cloudfocker/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	Describe("Parsing a ContainerConfig", func() {
		Context("with a staging config ", func() {
			It("should return a slice with all required arguments", func() {
				os.Setenv("CLOUDFOCKER_HOME", "/home/testuser/.cloudfocker")
				thisUser, _ := user.Current()
				userId := thisUser.Uid
				stageConfig := config.NewStageContainerConfig("/home/testuser/testapp", config.NewDirectories("/home/testuser/.cloudfocker"))
				parsedRunCommand := docker.ParseRunCommand(stageConfig)
				Expect(strings.Join(parsedRunCommand, " ")).To(Equal("-u=" + userId + " --name=cloudfocker-staging --volume=/home/testuser/.cloudfocker/buildpacks:/tmp/cloudfockerbuildpacks --volume=/home/testuser/.cloudfocker/cache:/tmp/cache --volume=/home/testuser/.cloudfocker/droplet:/tmp/droplet --volume=/home/testuser/.cloudfocker/focker:/focker --volume=/home/testuser/.cloudfocker/result:/tmp/result --volume=/home/testuser/testapp:/app cloudfocker-base:latest /focker/fock stage internal"))
			})
		})
		Context("with a runtime config ", func() {
			It("should return a slice with all required arguments", func() {
				os.Setenv("CLOUDFOCKER_HOME", "/home/testuser/.cloudfocker")
				thisUser, _ := user.Current()
				userId := thisUser.Uid
				testRuntimeContainerConfig := testRuntimeContainerConfig()
				parsedRunCommand := docker.ParseRunCommand(testRuntimeContainerConfig)
				Expect(strings.Join(parsedRunCommand, " ")).To(Equal("-u=" + userId + " --name=cloudfocker-runtime -d --volume=/home/testuser/testapp/app:/app --publish=8080:8080 --env=\"HOME=/app\" --env=\"PORT=8080\" --env=\"TMPDIR=/app/tmp\" cloudfocker-base:latest /bin/bash /app/cloudfocker-start.sh /app test test test"))
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
