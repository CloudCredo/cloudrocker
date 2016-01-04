package godocker_test

import (
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"github.com/cloudcredo/cloudrocker/config"
	"github.com/cloudcredo/cloudrocker/godocker"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	Describe("Parsing a ContainerConfig for a Docker run command", func() {
		Context("with a staging config ", func() {
			It("should return a slice with all required arguments", func() {
				os.Setenv("CLOUDROCKER_HOME", "/home/testuser/.cloudrocker")
				thisUser, _ := user.Current()
				userID := thisUser.Uid
				stageConfig := config.NewStageContainerConfig(config.NewDirectories("/home/testuser/.cloudrocker"))
				parsedRunCommand := godocker.ParseRunCommand(stageConfig)
				Expect(strings.Join(parsedRunCommand, " ")).To(Equal("-u=" + userID +
					" --name=cloudrocker-staging " +
					"--volume=/home/testuser/.cloudrocker/buildpacks:/cloudrockerbuildpacks " +
					"--volume=/home/testuser/.cloudrocker/rocker:/rocker " +
					"--volume=/home/testuser/.cloudrocker/staging:/tmp/app " +
					"--volume=/home/testuser/.cloudrocker/tmp:/tmp " +
					"--env=\"CF_STACK=cflinuxfs2\" " +
					"cloudrocker-base:latest " +
					"/rocker/rock stage internal"))
			})
		})

		Context("with a runtime config ", func() {
			It("should return a slice with all required arguments", func() {
				os.Setenv("CLOUDROCKER_HOME", "/home/testuser/.cloudrocker")
				thisUser, _ := user.Current()
				userID := thisUser.Uid
				testRuntimeContainerConfig := testRuntimeContainerConfig()
				parsedRunCommand := godocker.ParseRunCommand(testRuntimeContainerConfig)
				Expect(strings.Join(parsedRunCommand, " ")).To(Equal("-u=" + userID +
					" --name=cloudrocker-runtime -d " +
					"--volume=/home/testuser/testapp/app:/app " +
					"--publish=8080:8080 " +
					"--env=\"HOME=/app\" " +
					"--env=\"PORT=8080\" " +
					"--env=\"TMPDIR=/app/tmp\" " +
					"cloudrocker-base:latest " +
					"/bin/bash /app/cloudrocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh /app the start command \"quoted string with spaces\""))
			})
		})
	})

	Describe("Parsing a ContainerConfig for a Docker build command", func() {
		Context("with a runtime config ", func() {
			It("should write a valid Dockerfile", func() {
				tmpDropletDir, err := ioutil.TempDir(os.TempDir(), "parser-test-tmp-droplet")
				Expect(err).ShouldNot(HaveOccurred())
				testRuntimeContainerConfig := testRuntimeContainerConfig()
				testRuntimeContainerConfig.DropletDir = tmpDropletDir

				godocker.WriteRuntimeDockerfile(testRuntimeContainerConfig)

				expected, err := ioutil.ReadFile("fixtures/build/Dockerfile")
				Expect(err).ShouldNot(HaveOccurred())
				result, err := ioutil.ReadFile(tmpDropletDir + "/Dockerfile")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result).To(Equal(expected))

				os.RemoveAll(tmpDropletDir)
			})
		})

		Context("with a base image building config ", func() {
			It("should write a valid Dockerfile", func() {
				tmpBaseConfigDir, err := ioutil.TempDir(os.TempDir(), "parser-test-base-config")
				Expect(err).ShouldNot(HaveOccurred())
				testBaseConfigContainerConfig := testBaseConfigContainerConfig(tmpBaseConfigDir)

				godocker.WriteBaseImageDockerfile(testBaseConfigContainerConfig)

				result, err := ioutil.ReadFile(tmpBaseConfigDir + "/Dockerfile")
				Expect(err).ShouldNot(HaveOccurred())
				thisUser, _ := user.Current()
				userID := thisUser.Uid
				Expect(result).To(Equal([]byte(`FROM cloudrocker-raw:latest
RUN id vcap || /usr/sbin/useradd -mU -u ` + userID + ` -d /app -s /bin/bash vcap
RUN mkdir -p /app/tmp && chown -R vcap:vcap /app
`)))

				os.RemoveAll(tmpBaseConfigDir)
			})
		})
	})
})

func testRuntimeContainerConfig() (containerConfig *config.ContainerConfig) {
	containerConfig = &config.ContainerConfig{
		ContainerName:  "cloudrocker-runtime",
		SrcImageTag:    "cloudrocker-base:latest",
		PublishedPorts: map[int]int{8080: 8080},
		Mounts: map[string]string{
			"/home/testuser/testapp" + "/app": "/app",
		},
		Command: append([]string{"/bin/bash", "/app/cloudrocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh", "/app"},
			[]string{"the", "start", "command", `"quoted`, "string", "with", `spaces"`}...),
		Daemon: true,
		EnvVars: map[string]string{
			"HOME":          "/app",
			"TMPDIR":        "/app/tmp",
			"PORT":          "8080",
			"VCAP_SERVICES": "",
		},
	}
	return
}

func testBaseConfigContainerConfig(tmpBaseConfigDir string) (containerConfig *config.ContainerConfig) {
	containerConfig = &config.ContainerConfig{
		BaseConfigDir: tmpBaseConfigDir,
		SrcImageTag:   "cloudrocker-raw:latest",
		DstImageTag:   "cloudrocker-base:latest",
	}
	return
}
