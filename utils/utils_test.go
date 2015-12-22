package utils_test

import (
	"io/ioutil"
	"os"

	"github.com/cloudcredo/cloudrocker/utils"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Describe("Getting a rootfs URL", func() {
		Context("without a rootfs env var set", func() {
			It("should return the default URL", func() {
				os.Setenv("ROCKER_ROOTFS_URL", "")
				Expect(utils.GetRootfsUrl()).To(Equal("https://s3.amazonaws.com/blob.cfblob.com/978883d5-2e4d-495b-8aec-fc7c7e2988ad"))
			})
		})

		Context("with a rootfs env var set", func() {
			It("should return the specified URL", func() {
				os.Setenv("ROCKER_ROOTFS_URL", "dave")
				Expect(utils.GetRootfsUrl()).To(Equal("dave"))
			})
		})
	})

	Describe("Getting the CLOUDROCKER_HOME", func() {
		Context("without a CLOUDROCKER_HOME env var set", func() {
			It("should return the default URL", func() {
				os.Setenv("CLOUDROCKER_HOME", "")
				Expect(utils.CloudrockerHome()).To(Equal(os.Getenv("HOME") + "/cloudrocker"))
			})
		})

		Context("with a CLOUDROCKER_HOME env var set", func() {
			It("should return the specified URL", func() {
				os.Setenv("CLOUDROCKER_HOME", "/dave")
				Expect(utils.CloudrockerHome()).To(Equal("/dave"))
			})
		})
	})

	Describe("Finding the subdirectories in a directory", func() {
		It("should return a slice of found subdirectories", func() {
			parentDir, _ := ioutil.TempDir(os.TempDir(), "utils-test-subdirs")
			os.Mkdir(parentDir+"/dir1", 0755)
			os.Mkdir(parentDir+"/dir2", 0755)
			ioutil.WriteFile(parentDir+"/testfile", []byte("test"), 0644)
			dirs, err := utils.SubDirs(parentDir)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(dirs).Should(ContainElement("dir1"))
			Expect(dirs).Should(ContainElement("dir2"))
			Expect(dirs).ShouldNot(ContainElement("testfile"))
			os.RemoveAll(parentDir)
		})
	})

	Describe("Copying the rocker binary to its own directory", func() {
		It("should create a rocker subdirectory with the rock binary inside it", func() {
			cloudrockerHome, _ := ioutil.TempDir(os.TempDir(), "utils-test-cp-rocker")
			err := utils.CopyRockerBinaryToDir(cloudrockerHome + "/rocker")
			Expect(err).ShouldNot(HaveOccurred())
			rockerOwnDirFile, err := os.Open(cloudrockerHome + "/rocker")
			rockerOwnDirContents, err := rockerOwnDirFile.Readdirnames(0)
			Expect(rockerOwnDirContents, err).Should(ContainElement("rock"))
			info, _ := os.Stat(cloudrockerHome + "/rocker/rock")
			mode := info.Mode()
			Expect(mode).To(Equal(os.FileMode(0755)))
			os.RemoveAll(cloudrockerHome)
		})
	})

	Describe("Adding the launcher run script to a directory", func() {
		It("should create a script called cloudrocker-start... with expected contents", func() {
			appDir, _ := ioutil.TempDir(os.TempDir(), "utils-test-launcher")
			utils.AddLauncherRunScript(appDir)
			written, _ := ioutil.ReadFile(appDir + "/cloudrocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh")
			fixture, _ := ioutil.ReadFile("fixtures/cloudrocker-start.sh")
			Expect(written).To(Equal(fixture))
			os.RemoveAll(appDir)
		})
	})

	Describe("Getting the user's PWD", func() {
		It("should return the PWD", func() {
			testDir, _ := ioutil.TempDir(os.TempDir(), "utils-test-pwd")
			os.Chdir(testDir)
			rootedPath, _ := os.Getwd()
			Expect(utils.Pwd()).To(Equal(rootedPath))
		})
	})
})
