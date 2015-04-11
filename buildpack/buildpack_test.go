package buildpack_test

import (
	"io/ioutil"
	"os"

	"github.com/cloudcredo/cloudrocker/buildpack"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/gbytes"
)

var _ = Describe("Buildpack", func() {
	var (
		buffer *gbytes.Buffer
	)
	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
	})
	Describe("Adding a Buildpack", func() {
		It("should download the buildpack from the specified URL", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "crocker-buildpack-test")
			buildpack.Add(buffer, "https://github.com/hatofmonkeys/not-a-buildpack", buildpackDir)
			Eventually(buffer).Should(gbytes.Say(`Downloading buildpack...`))
			Eventually(buffer).Should(gbytes.Say(`Downloaded buildpack.`))
			contents, err := ioutil.ReadDir(buildpackDir + "/not-a-buildpack")
			Expect(contents, err).Should(HaveLen(3))
			os.RemoveAll(buildpackDir)
		})
	})
	Describe("Removing a Buildpack", func() {
		Context("with the buildpack", func() {
			It("should remove the buildpack from the buildpack directory", func() {
				buildpackDir, _ := ioutil.TempDir(os.TempDir(), "crocker-buildpack-remove-test")
				os.Mkdir(buildpackDir+"/testbuildpack", 0755)
				ioutil.WriteFile(buildpackDir+"/testbuildpack/testfile", []byte("test"), 0644)
				err := buildpack.Delete(buffer, "testbuildpack", buildpackDir)
				Expect(err).ShouldNot(HaveOccurred())
				contents, err := ioutil.ReadDir(buildpackDir)
				Expect(contents, err).Should(HaveLen(0))
				Eventually(buffer).Should(gbytes.Say(`Deleted buildpack.`))
				os.RemoveAll(buildpackDir)
			})
		})
		Context("without the buildpack", func() {
			It("should not return an error", func() {
				buildpackDir, _ := ioutil.TempDir(os.TempDir(), "crocker-buildpack-remove-test")
				err := buildpack.Delete(buffer, "testbuildpack", buildpackDir)
				Expect(err).ShouldNot(HaveOccurred())
				os.RemoveAll(buildpackDir)
			})
		})
	})
	Describe("Listing buildpacks", func() {
		Context("with buildpacks", func() {
			It("should list the buildpacks in the buildpack directory", func() {
				buildpackDir, _ := ioutil.TempDir(os.TempDir(), "crocker-buildpack-list-test")
				os.Mkdir(buildpackDir+"/testbuildpack", 0755)
				ioutil.WriteFile(buildpackDir+"/testbuildpack/testfile", []byte("test"), 0644)
				os.Mkdir(buildpackDir+"/testbuildpack2", 0755)
				ioutil.WriteFile(buildpackDir+"/testbuildpack2/testfile", []byte("test"), 0644)
				err := buildpack.List(buffer, buildpackDir)
				Expect(err).ShouldNot(HaveOccurred())
				Eventually(buffer).Should(gbytes.Say(`testbuildpack`))
				Eventually(buffer).Should(gbytes.Say(`testbuildpack2`))
				os.RemoveAll(buildpackDir)
			})
		})
		Context("without buildpacks", func() {
			It("should say there are no buildpacks installed", func() {
				buildpackDir, _ := ioutil.TempDir(os.TempDir(), "crocker-buildpack-list-test")
				err := buildpack.List(buffer, buildpackDir)
				Expect(err).ShouldNot(HaveOccurred())
				Eventually(buffer).Should(gbytes.Say(`No buildpacks installed`))
				os.RemoveAll(buildpackDir)
			})
		})
	})
	Describe("Checking for the presence of at least one buildpack", func() {
		Context("with one buildpack", func() {
			It("should return without error", func() {
				buildpackDir, _ := ioutil.TempDir(os.TempDir(), "crocker-buildpack-test-buildpack")
				os.Mkdir(buildpackDir+"/testbuildpack", 0755)
				err := buildpack.AtLeastOneBuildpackIn(buildpackDir)
				Expect(err).ShouldNot(HaveOccurred())
				os.RemoveAll(buildpackDir)
			})
		})
		Context("with no buildpacks", func() {
			It("should return an error", func() {
				buildpackDir, _ := ioutil.TempDir(os.TempDir(), "crocker-buildpack-test-buildpack")
				err := buildpack.AtLeastOneBuildpackIn(buildpackDir)
				Expect(err).Should(HaveOccurred())
				os.RemoveAll(buildpackDir)
			})
		})
	})
})
