package utils_test

import (
	"io/ioutil"
	"os"

	"github.com/cloudcredo/cloudfocker/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Describe("Getting a rootfs URL", func() {
		Context("without a rootfs env var set", func() {
			It("should return the default URL", func() {
				os.Setenv("FOCKER_ROOTFS_URL", "")
				Expect(utils.GetRootfsUrl()).To(Equal("https://s3.amazonaws.com/blob.cfblob.com/fee97b71-17d7-4fab-a5b0-69d4112521e6"))
			})
		})
		Context("with a rootfs env var set", func() {
			It("should return the specified URL", func() {
				os.Setenv("FOCKER_ROOTFS_URL", "dave")
				Expect(utils.GetRootfsUrl()).To(Equal("dave"))
			})
		})
	})
	Describe("Getting the CLOUDFOCKER_HOME", func() {
		Context("without a CLOUDFOCKER_HOME env var set", func() {
			It("should return the default URL", func() {
				os.Setenv("CLOUDFOCKER_HOME", "")
				Expect(utils.CloudfockerHome()).To(Equal(os.Getenv("HOME") + "/.cloudfocker"))
			})
		})
		Context("with a CLOUDFOCKER_HOME env var set", func() {
			It("should return the specified URL", func() {
				os.Setenv("CLOUDFOCKER_HOME", "/dave")
				Expect(utils.CloudfockerHome()).To(Equal("/dave"))
			})
		})
	})
	Describe("Creating and cleaning application directories", func() {
		Context("without a previously staged application", func() {
			It("should create the correct directory structure", func() {
				cloudfockerHome, _ := ioutil.TempDir(os.TempDir(), "utils-test-create-clean")
				err := utils.CreateAndCleanAppDirs(cloudfockerHome)
				Expect(err).ShouldNot(HaveOccurred())
				cloudfockerHomeFile, err := os.Open(cloudfockerHome)
				cloudfockerHomeContents, err := cloudfockerHomeFile.Readdirnames(0)
				Expect(cloudfockerHomeContents, err).Should(ContainElement("buildpacks"))
				Expect(cloudfockerHomeContents, err).Should(ContainElement("droplet"))
				Expect(cloudfockerHomeContents, err).Should(ContainElement("cache"))
				Expect(cloudfockerHomeContents, err).Should(ContainElement("result"))
				Expect(cloudfockerHomeContents, err).Should(ContainElement("staging"))
				os.RemoveAll(cloudfockerHome)
			})
		})
		Context("with a previously staged application", func() {
			It("should clean the directory structure appropriately", func() {
				cloudfockerHome, _ := ioutil.TempDir(os.TempDir(), "utils-test-create-clean")
				dirs := map[string]bool{"/buildpacks": false, "/droplet": true, "/cache": false, "/result": true, "/staging": true}
				for dir, _ := range dirs {
					os.MkdirAll(cloudfockerHome+dir, 0755)
					ioutil.WriteFile(cloudfockerHome+dir+"/testfile", []byte("test"), 0644)
				}
				err := utils.CreateAndCleanAppDirs(cloudfockerHome)
				Expect(err).ShouldNot(HaveOccurred())
				for dir, clean := range dirs {
					dirFile, err := os.Open(cloudfockerHome + dir)
					dirContents, err := dirFile.Readdirnames(0)
					if clean {
						Expect(dirContents, err).ShouldNot(ContainElement("testfile"))
					} else {
						Expect(dirContents, err).Should(ContainElement("testfile"))
					}
				}
				os.RemoveAll(cloudfockerHome)
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
	Describe("Copying the focker binary to its own directory", func() {
		It("should create a focker subdirectory with the fock binary inside it", func() {
			cloudfockerHome, _ := ioutil.TempDir(os.TempDir(), "utils-test-cp-focker")
			err := utils.CopyFockerBinaryToOwnDir(cloudfockerHome)
			Expect(err).ShouldNot(HaveOccurred())
			fockerOwnDirFile, err := os.Open(cloudfockerHome + "/focker")
			fockerOwnDirContents, err := fockerOwnDirFile.Readdirnames(0)
			Expect(fockerOwnDirContents, err).Should(ContainElement("fock"))
			info, _ := os.Stat(cloudfockerHome + "/focker/fock")
			mode := info.Mode()
			Expect(mode).To(Equal(os.FileMode(0755)))
			os.RemoveAll(cloudfockerHome)
		})
	})
	Describe("Adding the soldier run script to a directory", func() {
		It("should create a script called cloudfocker-start.sh with expected contents", func() {
			appDir, _ := ioutil.TempDir(os.TempDir(), "utils-test-soldier")
			utils.AddSoldierRunScript(appDir)
			written, _ := ioutil.ReadFile(appDir + "/cloudfocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh")
			fixture, _ := ioutil.ReadFile("fixtures/cloudfocker-start.sh")
			Expect(written).To(Equal(fixture))
			os.RemoveAll(appDir)
		})
	})
})
