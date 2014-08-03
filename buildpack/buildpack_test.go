package buildpack_test

import (
	"io/ioutil"
	"os"

	"github.com/hatofmonkeys/cloudfocker/buildpack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
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
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "cfocker-buildpack-test")
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
				buildpackDir, _ := ioutil.TempDir(os.TempDir(), "cfocker-buildpack-remove-test")
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
				buildpackDir, _ := ioutil.TempDir(os.TempDir(), "cfocker-buildpack-remove-test")
				err := buildpack.Delete(buffer, "testbuildpack", buildpackDir)
				Expect(err).ShouldNot(HaveOccurred())
				os.RemoveAll(buildpackDir)
			})
		})
	})
})
