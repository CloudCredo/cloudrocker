package utils_test

import (
	"io/ioutil"
	"os"

	"github.com/hatofmonkeys/cloudfocker/utils"

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
				Expect(utils.Cloudfockerhome()).To(Equal(os.Getenv("HOME") + "/.cloudfocker"))
			})
		})
		Context("with a CLOUDFOCKER_HOME env var set", func() {
			It("should return the specified URL", func() {
				os.Setenv("CLOUDFOCKER_HOME", "/dave")
				Expect(utils.Cloudfockerhome()).To(Equal("/dave"))
			})
		})
	})
	Describe("Creating and cleaning application directories", func() {
		Context("without a previously staged application", func() {
			It("should create the correct directory structure", func() {
				cloudfockerhome, _ := ioutil.TempDir(os.TempDir(), "utils-test-create-clean")
				err := utils.CreateAndCleanAppDirs(cloudfockerhome)
				Expect(err).ShouldNot(HaveOccurred())
				cloudfockerhomeFile, err := os.Open(cloudfockerhome)
				cloudfockerhomeContents, err := cloudfockerhomeFile.Readdirnames(0)
				Expect(cloudfockerhomeContents, err).Should(ContainElement("buildpacks"))
				Expect(cloudfockerhomeContents, err).Should(ContainElement("droplet"))
				Expect(cloudfockerhomeContents, err).Should(ContainElement("cache"))
				Expect(cloudfockerhomeContents, err).Should(ContainElement("result"))
				os.RemoveAll(cloudfockerhome)
			})
		})
		Context("with a previously staged application", func() {
			It("should clean the directory structure appropriately", func() {
				cloudfockerhome, _ := ioutil.TempDir(os.TempDir(), "utils-test-create-clean")
				dirs := map[string]bool{"/buildpacks": false, "/droplet": true, "/cache": false, "/result": true}
				for dir, _ := range dirs {
					os.MkdirAll(cloudfockerhome+dir, 0755)
					ioutil.WriteFile(cloudfockerhome+dir+"/testfile", []byte("test"), 0644)
				}
				err := utils.CreateAndCleanAppDirs(cloudfockerhome)
				Expect(err).ShouldNot(HaveOccurred())
				for dir, clean := range dirs {
					dirFile, err := os.Open(cloudfockerhome + dir)
					dirContents, err := dirFile.Readdirnames(0)
					if clean {
						Expect(dirContents, err).ShouldNot(ContainElement("testfile"))
					} else {
						Expect(dirContents, err).Should(ContainElement("testfile"))
					}
				}
				os.RemoveAll(cloudfockerhome)
			})
		})
	})
	Describe("Checking for the presence of at least one buildpack", func() {
		Context("with one buildpack", func() {
			It("should return without error", func() {
				buildpackDir, _ := ioutil.TempDir(os.TempDir(), "utils-test-buildpack")
				os.Mkdir(buildpackDir+"/testbuildpack", 0755)
				err := utils.AtLeastOneBuildpackIn(buildpackDir)
				Expect(err).ShouldNot(HaveOccurred())				
				os.RemoveAll(buildpackDir)
			})
		})
		Context("with no buildpacks", func() {
			It("should return an error", func() {
				buildpackDir, _ := ioutil.TempDir(os.TempDir(), "utils-test-buildpack")
				err := utils.AtLeastOneBuildpackIn(buildpackDir)
				Expect(err).Should(HaveOccurred())
				os.RemoveAll(buildpackDir)				
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
})
