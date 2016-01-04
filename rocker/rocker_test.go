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
		Context("REALDOCKER", func() {
			It("should tell Docker to output its version", func() {
				rocker.DockerVersion(buffer)
				Eventually(buffer).Should(gbytes.Say("Checking Docker version"))
				Eventually(buffer).Should(gbytes.Say("Client OS/Arch: "))
				Eventually(buffer).Should(gbytes.Say("Server Go version: go"))
			})
		})
	})

	Describe("Managing images", func() {
		Context("REALDOCKER", func() {
			Describe("Bootstrapping the raw image", func() {
				//This works, but speed depends on your net connection
				XIt("should download and tag the raw filesystem", func() {
					fmt.Println("Downloading rootfs - this could take a while")
					testrocker.ImportRootfsImage(buffer)
					Eventually(buffer, 600).Should(gbytes.Say(`[a-f0-9]{64}`))
					Eventually(buffer).Should(gbytes.Say(`Successfully built [a-f0-9]{12}`))
				})
			})

			Describe("Creating the base image", func() {
				It("should create a base image", func() {
					testrocker.BuildBaseImage(buffer)
					Eventually(buffer).Should(gbytes.Say(`Successfully built [a-f0-9]{12}`))
				})
			})
		})
	})

	Describe("Managing buildpacks", func() {
		var (
			buildpackDir string
		)

		BeforeEach(func() {
			buildpackDir, _ = ioutil.TempDir(os.TempDir(), "crocker-buildpack-test")
		})

		AfterEach(func() {
			os.RemoveAll(buildpackDir)
		})

		Describe("Adding a buildpack", func() {
			It("should download the buildpack and add it to the buildpack directory", func() {
				testrocker.AddBuildpack(buffer, "https://github.com/hatofmonkeys/not-a-buildpack", buildpackDir)
				Eventually(buffer).Should(gbytes.Say(`Downloading buildpack...`))
				Eventually(buffer, 10).Should(gbytes.Say(`Downloaded buildpack.`))
			})
		})

		Describe("Deleting a buildpack", func() {
			It("should delete the buildpack from the buildpack directory", func() {
				testrocker.AddBuildpack(buffer, "https://github.com/hatofmonkeys/not-a-buildpack", buildpackDir)
				testrocker.DeleteBuildpack(buffer, "not-a-buildpack")
				Eventually(buffer).Should(gbytes.Say(`Deleted buildpack.`))
			})
		})

		Describe("Listing buildpacks", func() {
			It("should list the buildpacks in the buildpack directory", func() {
				testrocker.AddBuildpack(buffer, "https://github.com/hatofmonkeys/not-a-buildpack", buildpackDir)
				testrocker.ListBuildpacks(buffer)
				Eventually(buffer).Should(gbytes.Say(`not-a-buildpack`))
			})
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
			Context("REALDOCKER", func() {
				BeforeEach(func() {
					cp("fixtures/stage/buildpacks", cloudrockerHome)
					originalDir = utils.Pwd()
					os.Chdir("fixtures/stage/apps/bash-app")
					testrocker = rocker.NewRocker()
					err := testrocker.RunStager(buffer)
					Expect(err).ShouldNot(HaveOccurred())
				})

				It("should transfer dotfiles to the staging directory", func() {
					stagingDir, err := os.Open(config.NewDirectories(cloudrockerHome).Staging())
					stagingDirContents, err := stagingDir.Readdirnames(0)
					Expect(stagingDirContents, err).Should(ContainElement(".testdotfile"))
				})

				It("should create the droplet", func() {
					dropletDir, err := os.Open(config.NewDirectories(cloudrockerHome).Tmp())
					dropletDirContents, err := dropletDir.Readdirnames(0)
					Expect(dropletDirContents, err).Should(ContainElement("droplet"))
				})
			})
		})

		Context("with a buildpack that doesn't detect", func() {
			Context("REALDOCKER", func() {
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
	})

	Describe("Managing applications", func() {
		Context("REALDOCKER", func() {
			var (
				cloudrockerHome string
				appDir          string
				originalDir     string
			)

			BeforeEach(func() {
				cloudrockerHome, _ = ioutil.TempDir(os.TempDir(), "rocker-runtime-test")
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
				Context("without a tag", func() {
					It("should output the built image ID", func() {
						testrocker.RunStager(buffer)
						testrocker.BuildRuntimeImage(buffer)
						Eventually(buffer).Should(gbytes.Say(`Successfully built [a-f0-9]{12}`))
					})
				})
				Context("with a tag", func() {
					It("should output the built image ID", func() {
						testrocker.RunStager(buffer)
						testrocker.BuildRuntimeImage(buffer, "rockertestsuite/image-tag:test")
						Eventually(buffer).Should(gbytes.Say(`Successfully built [a-f0-9]{12}`))
					})
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
					Expect(cloudrockerHomeContents, err).Should(ContainElement("droplet"))
					os.RemoveAll(cloudrockerHome)
				})
			})
			Context("with a previously staged application", func() {
				It("should clean the directory structure appropriately", func() {
					cloudrockerHome, _ := ioutil.TempDir(os.TempDir(), "utils-test-create-clean")
					dirs := map[string]bool{"/buildpacks": false, "/droplet": true, "/tmp/cache": false, "/staging": true}
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
