package docker_test

import (
	"os"
	"strings"

	"github.com/hatofmonkeys/cloudfocker/config"
	"github.com/hatofmonkeys/cloudfocker/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	Describe("Parsing a RunConfig", func() {
		It("should return a slice with all required arguments", func() {
			os.Setenv("CLOUDFOCKER_HOME", "/home/testuser/.cloudfocker")
			stageConfig := config.NewStageRunConfig("/home/testuser/testapp")
			parsedRunCommand := docker.ParseRunCommand(stageConfig)
			Expect(strings.Join(parsedRunCommand, " ")).To(Equal("--volume=/home/testuser/.cloudfocker/buildpacks:/tmp/cloudfockerbuildpacks --volume=/home/testuser/.cloudfocker/cache:/tmp/cache --volume=/home/testuser/.cloudfocker/droplet:/tmp/droplet --volume=/home/testuser/.cloudfocker/focker:/focker --volume=/home/testuser/.cloudfocker/result:/tmp/result --volume=/home/testuser/testapp:/app --name=cloudfocker-staging cloudfocker-base:latest /focker/fock stage"))
		})
	})
})
