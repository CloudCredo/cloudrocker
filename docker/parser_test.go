package docker_test

import (
	"os"
	"os/user"
	"strings"

	"github.com/hatofmonkeys/cloudfocker/config"
	"github.com/hatofmonkeys/cloudfocker/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	Describe("Parsing a RunConfig", func() {
		Context("with a staging config ", func() {
			It("should return a slice with all required arguments", func() {
				os.Setenv("CLOUDFOCKER_HOME", "/home/testuser/.cloudfocker")
				thisUser, _ := user.Current()
				userId := thisUser.Uid
				stageConfig := config.NewStageRunConfig("/home/testuser/testapp")
				parsedRunCommand := docker.ParseRunCommand(stageConfig)
				Expect(strings.Join(parsedRunCommand, " ")).To(Equal("--volume=/home/testuser/.cloudfocker/buildpacks:/tmp/cloudfockerbuildpacks --volume=/home/testuser/.cloudfocker/cache:/tmp/cache --volume=/home/testuser/.cloudfocker/droplet:/tmp/droplet --volume=/home/testuser/.cloudfocker/focker:/focker --volume=/home/testuser/.cloudfocker/result:/tmp/result --volume=/home/testuser/testapp:/app -u=" + userId + " --name=cloudfocker-staging cloudfocker-base:latest /focker/fock stage-internal"))
			})
		})
		Context("with a runtime config ", func() {
			It("should return a slice with all required arguments", func() {
				os.Setenv("CLOUDFOCKER_HOME", "/home/testuser/.cloudfocker")
				thisUser, _ := user.Current()
				userId := thisUser.Uid
				testStageConfig := testRuntimeRunConfig()
				parsedRunCommand := docker.ParseRunCommand(testStageConfig)
				Expect(strings.Join(parsedRunCommand, " ")).To(Equal("--volume=/home/testuser/testapp/app:/app -u=" + userId + " --name=cloudfocker-runtime -d --env=\"HOME=/app\" --env=\"PORT=8080\" --env=\"TMPDIR=/app/tmp\" cloudfocker-base:latest /bin/bash /app/cloudfocker-start.sh /app test test test"))
			})
		})
	})
})

func testRuntimeRunConfig() (runConfig *config.RunConfig) {
	runConfig = &config.RunConfig{
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
