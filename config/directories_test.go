package config_test

import (
	"io/ioutil"
	"os"

	"github.com/cloudcredo/cloudfocker/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Directories", func() {
	Describe("Provide a structure for directories", func() {
		It("should return the buildpacks directory", func() {
			cloudFockerHomeDir, _ := ioutil.TempDir(os.TempDir(), "utils-test-create-clean")
			testDirectories := config.NewDirectories(cloudFockerHomeDir)
			Expect(testDirectories.Buildpacks()).To(Equal(cloudFockerHomeDir + "/buildpacks"))
			os.RemoveAll(cloudFockerHomeDir)
		})
	})
})
