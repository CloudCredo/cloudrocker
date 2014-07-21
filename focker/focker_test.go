package focker_test

import (
	"fmt"

	"github.com/hatofmonkeys/cloudfocker/focker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
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
			testfocker.DockerVersion(buffer)
			Eventually(buffer).Should(gbytes.Say(`Checking Docker version`))
			Eventually(buffer).Should(gbytes.Say(`Client API version: `))
			Eventually(buffer).Should(gbytes.Say(`Go version \(client\): go`))
		})
	})

	Describe("Bootstrapping the base image", func() {
		//This works, but speed depends on your net connection
		XIt("should download and tag the lucid64 filesystem", func() {
			fmt.Println("Downloading lucid64 - this could take a while")
			testfocker.ImportRootfsImage(buffer)
			Eventually(buffer, 600).Should(gbytes.Say(`[a-f0-9]{64}`))
		})
	})

	Describe("Writing a dockerfile", func() {
		It("should write a valid dockerfile", func() {
			testfocker.WriteDockerfile(buffer)
			Eventually(buffer).Should(gbytes.Say(`FROM`))
		})
	})

	Describe("Building a docker image", func() {
		It("should output a built image tag", func() {
			testfocker.BuildImage(buffer)
			Eventually(buffer, 20).Should(gbytes.Say(`Successfully built [a-f0-9]{12}`))
		})
	})

	/*
	  Describe("Running the docker container", func() {
	  	It("should output a valid URL for the running application", func() {
	  		Expect(true).To(Equal(true))
	  		})
	  	})
	*/
})
