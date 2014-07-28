package dockerfile_test

import (
	"io/ioutil"
	"os"

	"github.com/hatofmonkeys/cloudfocker/config"
	"github.com/hatofmonkeys/cloudfocker/dockerfile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Dockerfile", func() {
	var (
		testdockerfile *dockerfile.Dockerfile
		buffer         *gbytes.Buffer
	)

	BeforeEach(func() {
		testdockerfile = dockerfile.NewDockerfile()
		buffer = gbytes.NewBuffer()
	})

	Describe("Getting an empty dockerfile", func() {
		It("should return an empty dockerfile", func() {
			Expect(len(testdockerfile.Commands)).To(Equal(0))
		})
	})
	Describe("Creating a dockerfile", func() {
		It("should populate the dockerfile information", func() {
			testdockerfile.Create()
			Expect(len(testdockerfile.Commands)).To(Equal(7))
		})
	})
	Describe("Creating a staging dockerfile", func() {
		It("should populate the dockerfile information", func() {
			testdockerfile.CreateStaging()
			Expect(len(testdockerfile.Commands)).To(Equal(3))
		})
	})
	Describe("Creating a configured dockerfile", func() {
		Context("when supplied a staging config", func() {
			It("should populate the staging dockerfile information", func() {
				testdockerfile.CreateFromConfig(config.NewStageBuildConfig())
				Expect(len(testdockerfile.Commands)).To(Equal(3))
				Expect(testdockerfile.Commands).Should(ContainElement("FROM cloudfocker-base:latest"))
				Expect(testdockerfile.Commands).Should(ContainElement("ADD fock /"))
				Expect(testdockerfile.Commands).Should(ContainElement("ENTRYPOINT [\"/fock\", \"stage\"]"))
			})
		})
		Context("when supplied a runtime config", func() {
			It("should populate the runtime dockerfile information", func() {

			})
		})
	})
	Describe("Writing a dockerfile", func() {
		It("should write the dockerfile to a writer", func() {
			testdockerfile.Create()
			testdockerfile.Write(buffer)
			Eventually(buffer).Should(gbytes.Say(`FROM`))
		})
	})
	Describe("Persisting a dockerfile", func() {
		It("should persist the dockerfile to a file", func() {
			file, _ := ioutil.TempFile("", "focker-testing")
			filename := file.Name()
			file.Close()
			testdockerfile.Create()
			err := testdockerfile.Persist(filename)
			Expect(err).ShouldNot(HaveOccurred())
			contents, _ := ioutil.ReadFile(filename)
			Expect(contents).ShouldNot(BeEmpty())
			os.Remove(filename)
		})
	})
})
