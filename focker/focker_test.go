package focker_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/hatofmonkeys/cloudfocker/focker"

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

	Describe("Running the docker container", func() {
		It("should output a valid URL for the running application", func() {
			testfocker.RunContainer(buffer)
			Eventually(buffer, 20).Should(gbytes.Say(`Successfully built [a-f0-9]{12}`))
			Eventually(buffer).Should(gbytes.Say(`[a-f0-9]{64}`))
			Eventually(buffer).Should(gbytes.Say(`Connect to your running application at http://localhost:8080/`))
			Eventually(statusCodeChecker).Should(Equal(200))
			testfocker.StopContainer(buffer)
		})
	})

	Describe("Stopping the docker container", func() {
		It("should output the stopped image ID, not respond to HTTP, and delete the container", func() {
			testfocker.RunContainer(buffer)
			testfocker.StopContainer(buffer)
			Eventually(buffer).Should(gbytes.Say(`Stopping the CloudFocker container...`))
			Eventually(buffer).Should(gbytes.Say(`cloudfocker-container`))
			Eventually(statusCodeChecker).Should(Equal(0))
			Eventually(buffer).Should(gbytes.Say(`Deleting the CloudFocker container...`))
			Eventually(buffer).Should(gbytes.Say(`cloudfocker-container`))
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

	Describe("Building an application droplet", func() {
		It("should run the buildpack runner from linux-circus", func() {
			buildpackDir, _ := ioutil.TempDir(os.TempDir(), "cfocker-runner-test")
			err := testfocker.StageApp(buffer, buildpackDir)
			Expect(err).Should(MatchError("no valid buildpacks detected"))
			Eventually(buffer).Should(gbytes.Say(`Running Buildpacks...`))
		})
	})

	Describe("Staging an application", func() {
		It("should populate the droplet directory", func() {
			cloudfockerHome, _ := ioutil.TempDir(os.TempDir(), "focker-staging-test")
			os.Setenv("CLOUDFOCKER_HOME", cloudfockerHome)
			cp("fixtures/buildpacks", cloudfockerHome)
			testfocker.RunStager(buffer, "fixtures/apps/bash-app")
			dropletDir, err := os.Open(cloudfockerHome + "/droplet")
			dropletDirContents, err := dropletDir.Readdirnames(0)
			Expect(dropletDirContents, err).Should(ContainElement("app"))
			Expect(dropletDirContents, err).Should(ContainElement("logs"))
			Expect(dropletDirContents, err).Should(ContainElement("staging_info.yml"))
			Expect(dropletDirContents, err).Should(ContainElement("tmp"))
			os.RemoveAll(cloudfockerHome)
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
