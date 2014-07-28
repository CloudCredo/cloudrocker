package utils_test

import (
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
})
