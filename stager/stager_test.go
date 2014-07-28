package stager_test

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry-incubator/linux-circus/buildpackrunner"
	"github.com/hatofmonkeys/cloudfocker/stager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

type TestRunner struct {
	RunCalled bool
}

func (f *TestRunner) Run() error {
	f.RunCalled = true
	return nil
}

var _ = Describe("Stager", func() {
	Describe("Running a buildpack", func() {
		It("should tell a buildpack runner to run", func() {
			buffer := gbytes.NewBuffer()
			testrunner := new(TestRunner)
			err := stager.RunBuildpack(buffer, testrunner)
			Expect(testrunner.RunCalled).To(Equal(true))
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(buffer).Should(gbytes.Say(`Running Buildpacks...`))
		})
	})

	Describe("Getting a buildpack runner", func() {
		It("should return the address of a valid buildpack runner, with a correct buildpack list", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "cfocker-buildpackrunner-test")
			os.Mkdir(buildpackDir+"/test-buildpack", 0755)
			ioutil.WriteFile(buildpackDir+"/test-buildpack"+"/testfile", []byte("test"), 0644)
			runner := stager.NewBuildpackRunner(buildpackDir)
			var runnerVar *buildpackrunner.Runner
			Expect(runner).Should(BeAssignableToTypeOf(runnerVar))
			md5BuildpackName := fmt.Sprintf("%x", md5.Sum([]byte("test-buildpack")))
			md5BuildpackDir, err := os.Open("/tmp/buildpacks")
			contents, err := md5BuildpackDir.Readdirnames(0)
			Expect(contents, err).Should(ContainElement(md5BuildpackName))
			md5Buildpack, err := os.Open(md5BuildpackDir.Name() + "/" + md5BuildpackName)
			buildpackContents, err := md5Buildpack.Readdirnames(0)
			Expect(buildpackContents, err).Should(ContainElement("testfile"))
			os.RemoveAll(buildpackDir)
			os.RemoveAll("/tmp/buildpacks")
		})
	})
})
