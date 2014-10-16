package stager_test

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudcredo/cloudfocker/config"
	"github.com/cloudcredo/cloudfocker/stager"
	"github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/cloudfoundry-incubator/linux-circus/buildpackrunner"

	. "github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudfocker/Godeps/_workspace/src/github.com/onsi/gomega/gbytes"
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

	Describe("Validating a staged application", func() {
		Context("with something that looks like a staged application", func() {
			It("should not return an error", func() {
				cfhome, _ := ioutil.TempDir(os.TempDir(), "stager-test-staged")
				dropletDir := config.NewDirectories(cfhome).Droplet()
				os.MkdirAll(dropletDir+"/app", 0755)
				ioutil.WriteFile(dropletDir+"/staging_info.yml", []byte("test-staging-info"), 0644)
				err := stager.ValidateStagedApp(config.NewDirectories(cfhome))
				Expect(err).ShouldNot(HaveOccurred())
				os.RemoveAll(cfhome)
			})
		})
		Context("without something that looks like a staged application", func() {
			Context("because we have no droplet", func() {
				It("should return an error about a missing droplet", func() {
					cfhome, _ := ioutil.TempDir(os.TempDir(), "stager-test-staged")
					err := stager.ValidateStagedApp(config.NewDirectories(cfhome))
					Expect(err).Should(MatchError("Staging failed - have you added a buildpack for this type of application?"))
					os.RemoveAll(cfhome)
				})
			})
			Context("because we have no staging_info.yml", func() {
				It("should return an error about missing staging info", func() {
					cfhome, _ := ioutil.TempDir(os.TempDir(), "stager-test-staged")
					os.MkdirAll(cfhome+"/tmp/droplet/app", 0755)
					err := stager.ValidateStagedApp(config.NewDirectories(cfhome))
					Expect(err).Should(MatchError("Staging failed - no staging info was produced by the matching buildpack!"))
					os.RemoveAll(cfhome)
				})
			})
		})
	})
})
