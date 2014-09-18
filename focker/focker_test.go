package focker_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/cloudcredo/cloudfocker/config"
	"github.com/cloudcredo/cloudfocker/focker"
	"github.com/cloudcredo/cloudfocker/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Focker", func() {
	var (
		testfocker *focker.Focker
		buffer     *gbytes.Buffer
	)
	BeforeEach(func() {
		testfocker = focker.NewFocker()
		buffer = gbytes.NewBuffer()
	})

	Describe("Displaying the docker version", func() {
		It("should tell Docker to output its version", func() {
			focker.DockerVersion(buffer)
			Eventually(buffer).Should(gbytes.Say(`Checking Docker version`))
			Eventually(buffer).Should(gbytes.Say(`Client API version: `))
			Eventually(buffer).Should(gbytes.Say(`Go version \(client\): go`))
		})
	})

	Describe("Bootstrapping the base image", func() {
		//This works, but speed depends on your net connection
		XIt("should download and tag the lucid64 filesystem", func() {
			fmt.Println("Downloading lucid64 - this could take a while")
			focker.ImportRootfsImage(buffer)
			Eventually(buffer, 600).Should(gbytes.Say(`[a-f0-9]{64}`))
		})
	})

	Describe("Adding a buildpack", func() {
		It("should download the buildpack and add it to the buildpack directory", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "cfocker-buildpack-test")
			testfocker.AddBuildpack(buffer, "https://github.com/hatofmonkeys/not-a-buildpack", buildpackDir)
			Eventually(buffer).Should(gbytes.Say(`Downloading buildpack...`))
			Eventually(buffer, 10).Should(gbytes.Say(`Downloaded buildpack.`))
			os.RemoveAll(buildpackDir)
		})
	})

	Describe("Deleting a buildpack", func() {
		It("should delete the buildpack from the buildpack directory", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "cfocker-buildpack-test")
			testfocker.AddBuildpack(buffer, "https://github.com/hatofmonkeys/not-a-buildpack", buildpackDir)
			testfocker.DeleteBuildpack(buffer, "not-a-buildpack")
			Eventually(buffer).Should(gbytes.Say(`Deleted buildpack.`))
			os.RemoveAll(buildpackDir)
		})
	})

	Describe("Listing buildpacks", func() {
		It("should list the buildpacks in the buildpack directory", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "cfocker-buildpack-test")
			testfocker.AddBuildpack(buffer, "https://github.com/hatofmonkeys/not-a-buildpack", buildpackDir)
			testfocker.ListBuildpacks(buffer)
			Eventually(buffer).Should(gbytes.Say(`not-a-buildpack`))
			os.RemoveAll(buildpackDir)
		})
	})

	Describe("Building an application droplet", func() {
		It("should run the buildpack runner from linux-circus", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "cfocker-runner-test")
			err := focker.StageApp(buffer, buildpackDir)
			Expect(err).Should(MatchError("no valid buildpacks detected"))
			Eventually(buffer).Should(gbytes.Say(`Running Buildpacks...`))
			os.RemoveAll(buildpackDir)
		})
	})

	Describe("Staging an application", func() {
		var (
			cloudfockerHome string
			originalDir     string
		)

		BeforeEach(func() {
			cloudfockerHome, _ = ioutil.TempDir(os.TempDir(), "focker-staging-test")
			os.Setenv("CLOUDFOCKER_HOME", cloudfockerHome)
		})

		AfterEach(func() {
			os.RemoveAll(cloudfockerHome)
			os.Chdir(originalDir)
		})

		Context("with a detected buildpack", func() {
			It("should populate the droplet directory", func() {
				cp("fixtures/stage/buildpacks", cloudfockerHome)
				originalDir = utils.Pwd()
				os.Chdir("fixtures/stage/apps/bash-app")
				testfocker = focker.NewFocker()
				err := testfocker.RunStager(buffer)
				Expect(err).ShouldNot(HaveOccurred())
				dropletDir, err := os.Open(config.NewDirectories(cloudfockerHome).Droplet())
				dropletDirContents, err := dropletDir.Readdirnames(0)
				Expect(dropletDirContents, err).Should(ContainElement("app"))
				Expect(dropletDirContents, err).Should(ContainElement("logs"))
				Expect(dropletDirContents, err).Should(ContainElement("staging_info.yml"))
				Expect(dropletDirContents, err).Should(ContainElement("tmp"))
			})
		})
		Context("with a buildpack that doesn't detect", func() {
			It("tell us we don't have a valid buildpack", func() {
				cp("fixtures/runtime/buildpacks", cloudfockerHome)
				originalDir = utils.Pwd()
				os.Chdir("fixtures/stage/apps/bash-app")
				testfocker = focker.NewFocker()
				err := testfocker.RunStager(buffer)
				Expect(err).Should(MatchError("Staging failed - have you added a buildpack for this type of application?"))
			})
		})
	})

	Describe("Managing applications", func() {
		var (
			cloudfockerHome string
			appDir          string
			originalDir     string
		)

		BeforeEach(func() {
			cloudfockerHome, _ = ioutil.TempDir(os.TempDir(), "focker-staging-test")
			os.Setenv("CLOUDFOCKER_HOME", cloudfockerHome)
			cp("fixtures/runtime/buildpacks", cloudfockerHome)

			appDir, _ = ioutil.TempDir(os.TempDir(), "focker-test-app")
			cp("fixtures/runtime/apps/cf-test-buildpack-app", appDir)

			originalDir = utils.Pwd()
			os.Chdir(appDir + "/cf-test-buildpack-app")

			testfocker = focker.NewFocker()
			testfocker.RunStager(buffer)
		})

		AfterEach(func() {
			os.RemoveAll(cloudfockerHome)
			os.RemoveAll(appDir)
			os.Chdir(originalDir)
		})

		Describe("when running an application", func() {
			Context("without a currently running application", func() {
				It("should output a valid URL for the running application", func() {
					testfocker.RunRuntime(buffer)
					Eventually(buffer).Should(gbytes.Say(`Connect to your running application at http://localhost:8080/`))
					Eventually(statusCodeChecker).Should(Equal(200))
					testfocker.StopRuntime(buffer)
				})
			})
			Context("with a currently running application", func() {
				It("should delete the current container and output a valid URL for the new running application", func() {
					testfocker.RunRuntime(buffer)
					Eventually(buffer).Should(gbytes.Say(`Connect to your running application at http://localhost:8080/`))
					Eventually(statusCodeChecker).Should(Equal(200))
					testfocker.RunRuntime(buffer)
					Consistently(buffer).ShouldNot(gbytes.Say(`Conflict`))
					Eventually(statusCodeChecker).Should(Equal(200))
				})
			})
		})

		Describe("when stopping a running an application", func() {
			It("should stop the application", func() {
				testfocker.RunStager(buffer)
				testfocker.RunRuntime(buffer)
				Eventually(statusCodeChecker).Should(Equal(200))
				testfocker.StopRuntime(buffer)
				Eventually(statusCodeChecker).Should(Equal(0))
			})
		})
	})
	Describe("Creating and cleaning application directories", func() {
		Context("without a previously staged application", func() {
			It("should create the correct directory structure", func() {
				cloudfockerHome, _ := ioutil.TempDir(os.TempDir(), "utils-test-create-clean")
				err := focker.CreateAndCleanAppDirs(config.NewDirectories(cloudfockerHome))
				Expect(err).ShouldNot(HaveOccurred())
				cloudfockerHomeFile, err := os.Open(cloudfockerHome)
				cloudfockerHomeContents, err := cloudfockerHomeFile.Readdirnames(0)
				Expect(cloudfockerHomeContents, err).Should(ContainElement("buildpacks"))
				Expect(cloudfockerHomeContents, err).Should(ContainElement("droplet"))
				Expect(cloudfockerHomeContents, err).Should(ContainElement("cache"))
				Expect(cloudfockerHomeContents, err).Should(ContainElement("result"))
				Expect(cloudfockerHomeContents, err).Should(ContainElement("staging"))
				os.RemoveAll(cloudfockerHome)
			})
		})
		Context("with a previously staged application", func() {
			It("should clean the directory structure appropriately", func() {
				cloudfockerHome, _ := ioutil.TempDir(os.TempDir(), "utils-test-create-clean")
				dirs := map[string]bool{"/buildpacks": false, "/droplet": true, "/cache": false, "/result": true, "/staging": true}
				for dir, _ := range dirs {
					os.MkdirAll(cloudfockerHome+dir, 0755)
					ioutil.WriteFile(cloudfockerHome+dir+"/testfile", []byte("test"), 0644)
				}
				err := focker.CreateAndCleanAppDirs(config.NewDirectories(cloudfockerHome))
				Expect(err).ShouldNot(HaveOccurred())
				for dir, clean := range dirs {
					dirFile, err := os.Open(cloudfockerHome + dir)
					dirContents, err := dirFile.Readdirnames(0)
					if clean {
						Expect(dirContents, err).ShouldNot(ContainElement("testfile"))
					} else {
						Expect(dirContents, err).Should(ContainElement("testfile"))
					}
				}
				os.RemoveAll(cloudfockerHome)
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
