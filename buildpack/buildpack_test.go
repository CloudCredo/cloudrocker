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
	Describe("Adding a Buildpack", func() {
		It("should download the buildpack from the specified URL", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "cfocker-buildpack-test")
			buffer := gbytes.NewBuffer()
			buildpack.Add(buffer, "https://github.com/hatofmonkeys/ruby-buildpack", buildpackDir)
			Eventually(buffer).Should(gbytes.Say(`Downloading buildpack...`))
			Eventually(buffer).Should(gbytes.Say(`Downloaded buildpack.`))
			contents, err := ioutil.ReadDir(buildpackDir + "/ruby-buildpack")
			Expect(contents, err).Should(HaveLen(23))
			os.RemoveAll(buildpackDir)
		})
	})
})
