package rocker_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/cloudcredo/cloudrocker/config"
	"github.com/cloudcredo/cloudrocker/rocker"
	"github.com/cloudcredo/cloudrocker/utils"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gbytes"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gexec"
)

var _ = Describe("Rocker", func() {
	var (
		testrocker *rocker.Rocker
		buffer     *gbytes.Buffer
	)
	BeforeEach(func() {
		testrocker = rocker.NewRocker()
		buffer = gbytes.NewBuffer()
	})

	Describe("Displaying the docker version", func() {
		It("should tell Docker to output its version", func() {
			rocker.DockerVersion(buffer)
			Eventually(buffer).Should(gbytes.Say(`Checking Docker version`))
			Eventually(buffer).Should(gbytes.Say(`Client API version: `))
			Eventually(buffer).Should(gbytes.Say(`Go version \(client\): go`))
		})
	})

	Describe("Bootstrapping the base image", func() {
		//This works, but speed depends on your net connection
		XIt("should download and tag the lucid64 filesystem", func() {
			fmt.Println("Downloading lucid64 - this could take a while")
			rocker.ImportRootfsImage(buffer)
			Eventually(buffer, 600).Should(gbytes.Say(`[a-f0-9]{64}`))
		})
	})

	Describe("Adding a buildpack", func() {
		It("should download the buildpack and add it to the buildpack directory", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "crocker-buildpack-test")
			testrocker.AddBuildpack(buffer, "https://github.com/hatofmonkeys/not-a-buildpack", buildpackDir)
			Eventually(buffer).Should(gbytes.Say(`Downloading buildpack...`))
			Eventually(buffer, 10).Should(gbytes.Say(`Downloaded buildpack.`))
			os.RemoveAll(buildpackDir)
		})
	})

	Describe("Deleting a buildpack", func() {
		It("should delete the buildpack from the buildpack directory", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "crocker-buildpack-test")
			testrocker.AddBuildpack(buffer, "https://github.com/hatofmonkeys/not-a-buildpack", buildpackDir)
			testrocker.DeleteBuildpack(buffer, "not-a-buildpack")
			Eventually(buffer).Should(gbytes.Say(`Deleted buildpack.`))
			os.RemoveAll(buildpackDir)
		})
	})

	Describe("Listing buildpacks", func() {
		It("should list the buildpacks in the buildpack directory", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "crocker-buildpack-test")
			testrocker.AddBuildpack(buffer, "https://github.com/hatofmonkeys/not-a-buildpack", buildpackDir)
			testrocker.ListBuildpacks(buffer)
			Eventually(buffer).Should(gbytes.Say(`not-a-buildpack`))
			os.RemoveAll(buildpackDir)
		})
	})

	Describe("Building an application droplet", func() {
		It("should run the buildpack runner from linux-circus", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "crocker-runner-test")
			err := testrocker.StageApp(buffer, buildpackDir)
			Expect(err).Should(MatchError("no valid buildpacks detected"))
			Eventually(buffer).Should(gbytes.Say(`Running Buildpacks...`))
			os.RemoveAll(buildpackDir)
		})
	})

	Describe("Staging an application", func() {
		var (
			cloudrockerHome string
			originalDir     string
		)

		BeforeEach(func() {
			cloudrockerHome, _ = ioutil.TempDir(os.TempDir(), "rocker-staging-test")
			os.Setenv("CLOUDROCKER_HOME", cloudrockerHome)
		})

		AfterEach(func() {
			os.RemoveAll(cloudrockerHome)
			os.Chdir(originalDir)
		})

		Context("with a detected buildpack", func() {
			It("should populate the droplet directory", func() {
				cp("fixtures/stage/buildpacks", cloudrockerHome)
				originalDir = utils.Pwd()
				os.Chdir("fixtures/stage/apps/bash-app")
				testrocker = rocker.NewRocker()
				err := testrocker.RunStager(buffer)
				Expect(err).ShouldNot(HaveOccurred())
				dropletDir, err := os.Open(config.NewDirectories(cloudrockerHome).Droplet())
				dropletDirContents, err := dropletDir.Readdirnames(0)
				Expect(dropletDirContents, err).Should(ContainElement("app"))
				Expect(dropletDirContents, err).Should(ContainElement("logs"))
				Expect(dropletDirContents, err).Should(ContainElement("staging_info.yml"))
				Expect(dropletDirContents, err).Should(ContainElement("tmp"))
			})
		})
		Context("with a buildpack that doesn't detect", func() {
			It("tell us we don't have a valid buildpack", func() {
				cp("fixtures/runtime/buildpacks", cloudrockerHome)
				originalDir = utils.Pwd()
				os.Chdir("fixtures/stage/apps/bash-app")
				testrocker = rocker.NewRocker()
				err := testrocker.RunStager(buffer)
				Expect(err).Should(MatchError("Staging failed - have you added a buildpack for this type of application?"))
			})
		})
	})

	Describe("Managing applications", func() {
		var (
			cloudrockerHome string
			appDir          string
			originalDir     string
		)

		BeforeEach(func() {
			cloudrockerHome, _ = ioutil.TempDir(os.TempDir(), "rocker-staging-test")
			os.Setenv("CLOUDROCKER_HOME", cloudrockerHome)
			cp("fixtures/runtime/buildpacks", cloudrockerHome)

			appDir, _ = ioutil.TempDir(os.TempDir(), "rocker-test-app")
			cp("fixtures/runtime/apps/cf-test-buildpack-app", appDir)

			originalDir = utils.Pwd()
			os.Chdir(appDir + "/cf-test-buildpack-app")

			testrocker = rocker.NewRocker()
			testrocker.RunStager(buffer)
		})

		AfterEach(func() {
			os.RemoveAll(cloudrockerHome)
			os.RemoveAll(appDir)
			os.Chdir(originalDir)
		})

		Describe("when running an application", func() {
			Context("without a currently running application", func() {
				It("should output a valid URL for the running application", func() {
					testrocker.RunRuntime(buffer)
					Eventually(buffer).Should(gbytes.Say(`Connect to your running application at http://localhost:8080/`))
					Eventually(statusCodeChecker).Should(Equal(200))
					testrocker.StopRuntime(buffer)
				})
			})
			Context("with a currently running application", func() {
				It("should delete the current container and output a valid URL for the new running application", func() {
					testrocker.RunRuntime(buffer)
					Eventually(buffer).Should(gbytes.Say(`Connect to your running application at http://localhost:8080/`))
					Eventually(statusCodeChecker).Should(Equal(200))
					testrocker.RunRuntime(buffer)
					Consistently(buffer).ShouldNot(gbytes.Say(`Conflict`))
					Eventually(statusCodeChecker).Should(Equal(200))
				})
			})
		})

		Describe("when stopping a running an application", func() {
			It("should stop the application", func() {
				testrocker.RunStager(buffer)
				testrocker.RunRuntime(buffer)
				Eventually(statusCodeChecker).Should(Equal(200))
				testrocker.StopRuntime(buffer)
				Eventually(statusCodeChecker).Should(Equal(0))
			})
		})
		Describe("when outputting a runnable application as a docker image", func() {
			It("should output the built image ID", func() {
				testrocker.RunStager(buffer)
				testrocker.BuildRuntimeImage(buffer)
				Eventually(buffer).Should(gbytes.Say(`Successfully built [a-f0-9]{12}`))
			})
		})
	})
	Describe("Creating and cleaning application directories", func() {
		Context("without a previously staged application", func() {
			It("should create the correct directory structure", func() {
				cloudrockerHome, _ := ioutil.TempDir(os.TempDir(), "utils-test-create-clean")
				err := rocker.CreateAndCleanAppDirs(config.NewDirectories(cloudrockerHome))
				Expect(err).ShouldNot(HaveOccurred())
				cloudrockerHomeFile, err := os.Open(cloudrockerHome)
				cloudrockerHomeContents, err := cloudrockerHomeFile.Readdirnames(0)
				Expect(cloudrockerHomeContents, err).Should(ContainElement("buildpacks"))
				Expect(cloudrockerHomeContents, err).Should(ContainElement("tmp"))
				Expect(cloudrockerHomeContents, err).Should(ContainElement("staging"))
				cloudrockerHomeTmpFile, err := os.Open(cloudrockerHome + "/tmp")
				cloudrockerHomeTmpContents, err := cloudrockerHomeTmpFile.Readdirnames(0)
				Expect(cloudrockerHomeTmpContents, err).Should(ContainElement("droplet"))
				Expect(cloudrockerHomeTmpContents, err).Should(ContainElement("result"))
				Expect(cloudrockerHomeTmpContents, err).Should(ContainElement("cache"))
				os.RemoveAll(cloudrockerHome)
			})
		})
		Context("with a previously staged application", func() {
			It("should clean the directory structure appropriately", func() {
				cloudrockerHome, _ := ioutil.TempDir(os.TempDir(), "utils-test-create-clean")
				dirs := map[string]bool{"/buildpacks": false, "/tmp/droplet": true, "/tmp/cache": false, "/tmp/result": true, "/staging": true}
				for dir, _ := range dirs {
					os.MkdirAll(cloudrockerHome+dir, 0755)
					ioutil.WriteFile(cloudrockerHome+dir+"/testfile", []byte("test"), 0644)
				}
				err := rocker.CreateAndCleanAppDirs(config.NewDirectories(cloudrockerHome))
				Expect(err).ShouldNot(HaveOccurred())
				for dir, clean := range dirs {
					dirFile, err := os.Open(cloudrockerHome + dir)
					dirContents, err := dirFile.Readdirnames(0)
					if clean {
						Expect(dirContents, err).ShouldNot(ContainElement("testfile"))
					} else {
						Expect(dirContents, err).Should(ContainElement("testfile"))
					}
				}
				os.RemoveAll(cloudrockerHome)
			})
		})
	})
})

func statusCodeChecker() int {
	res, err := http.Get("http://localhost:8080/")
	if err != nil {
		return 0
	} else {
		return res.StatusCode
	}
}

func cp(src string, dst string) {
	session, err := gexec.Start(
		exec.Command("cp", "-a", src, dst),
		GinkgoWriter,
		GinkgoWriter,
	)
	Ω(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
}
