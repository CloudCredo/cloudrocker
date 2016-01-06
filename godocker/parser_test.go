package godocker_test

import (
	"io/ioutil"
	"os"
	"os/user"

	"github.com/cloudcredo/cloudrocker/config"
	"github.com/cloudcredo/cloudrocker/godocker"

	goDockerClient "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	Describe("Parsing a ContainerConfig for a create container request", func() {
		Context("with a staging config", func() {
			It("should return a CreateContainerOptions with all required attributes", func() {
				os.Setenv("CLOUDROCKER_HOME", "/home/testuser/.cloudrocker")
				thisUser, _ := user.Current()
				userID := thisUser.Uid
				stageConfig := config.NewStageContainerConfig(config.NewDirectories("/home/testuser/.cloudrocker"))

				createContainerOptions := godocker.ParseCreateContainerOptions(stageConfig)

				Expect(createContainerOptions.Name).To(Equal("cloudrocker-staging"))
				Expect(createContainerOptions.Config.User).To(Equal(userID))
				Expect(createContainerOptions.Config.Env).To(Equal([]string{"CF_STACK=cflinuxfs2"}))
				Expect(createContainerOptions.Config.Image).To(Equal("cloudrocker-base:latest"))
				Expect(createContainerOptions.Config.Cmd).To(Equal([]string{"/rocker/rock", "stage", "internal"}))
				var binds = []string{
					"/home/testuser/.cloudrocker/buildpacks:/cloudrockerbuildpacks",
					"/home/testuser/.cloudrocker/rocker:/rocker",
					"/home/testuser/.cloudrocker/staging:/tmp/app",
					"/home/testuser/.cloudrocker/tmp:/tmp",
				}
				Expect(createContainerOptions.HostConfig.Binds).To(Equal(binds))
			})
		})

		Context("with a runtime config ", func() {
			It("should return a CreateContainerOptions with all required attributes", func() {
				thisUser, _ := user.Current()
				userID := thisUser.Uid
				testRuntimeContainerConfig := testRuntimeContainerConfig()

				createContainerOptions := godocker.ParseCreateContainerOptions(testRuntimeContainerConfig)

				Expect(createContainerOptions.Name).To(Equal("cloudrocker-runtime"))
				Expect(createContainerOptions.Config.User).To(Equal(userID))
				Expect(createContainerOptions.Config.Env).To(Equal([]string{
					"HOME=/app",
					"PORT=8080",
					"TMPDIR=/app/tmp",
				}))
				Expect(createContainerOptions.Config.Image).To(Equal("cloudrocker-base:latest"))
				Expect(createContainerOptions.Config.Cmd).To(Equal([]string{
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
				Expect(createContainerOptions.HostConfig.Binds).To(Equal([]string{
					"/home/testuser/testapp/app:/app",
				}))
				var portBindings = map[goDockerClient.Port][]goDockerClient.PortBinding{
					"8080/tcp": []goDockerClient.PortBinding{
						{
							HostPort: "8080",
						},
					},
				}
				Expect(createContainerOptions.HostConfig.PortBindings).To(Equal(portBindings))
			})
		})
	})

	Describe("Parsing a ContainerConfig for a Docker build command", func() {
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
