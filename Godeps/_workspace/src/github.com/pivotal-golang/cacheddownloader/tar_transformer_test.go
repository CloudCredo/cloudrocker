package cacheddownloader_test

import (
	"archive/tar"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/pivotal-golang/archiver/extractor/test_helper"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/pivotal-golang/cacheddownloader"
)

var _ = Describe("TarTransformer", func() {
	var (
		scratch string

		sourcePath      string
		destinationPath string

		transformedSize int64
		transformErr    error
	)

	archiveFiles := []test_helper.ArchiveFile{
		{Name: "some-file", Body: "some-contents"},
		{Name: "some-symlink", Link: "some-symlink-target"},
		{Name: "some-symlink-target", Body: "some-other-contents"},
	}

	verifyTarFile := func(path string) {
		file, err := os.Open(path)
		Ω(err).ShouldNot(HaveOccurred())

		tr := tar.NewReader(file)

		entry, err := tr.Next()
		Ω(err).ShouldNot(HaveOccurred())

		Ω(entry.Name).Should(Equal("some-file"))
		Ω(entry.Size).Should(Equal(int64(len("some-contents"))))
	}

	BeforeEach(func() {
		var err error

		scratch, err = ioutil.TempDir("", "tar-transformer-scratch")
		Ω(err).ShouldNot(HaveOccurred())

		destinationFile, err := ioutil.TempFile("", "destination")
		Ω(err).ShouldNot(HaveOccurred())

		err = destinationFile.Close()
		Ω(err).ShouldNot(HaveOccurred())

		destinationPath = destinationFile.Name()
	})

	AfterEach(func() {
		err := os.RemoveAll(scratch)
		Ω(err).ShouldNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		transformedSize, transformErr = TarTransform(sourcePath, destinationPath)
	})

	Context("when the file is already a .tar", func() {
		BeforeEach(func() {
			sourcePath = filepath.Join(scratch, "file.tar")

			test_helper.CreateTarArchive(sourcePath, archiveFiles)
		})

		It("renames the file to the destination", func() {
			verifyTarFile(destinationPath)
		})

		It("removes the source file", func() {
			_, err := os.Stat(sourcePath)
			Ω(err).Should(HaveOccurred())
		})

		It("returns its size", func() {
			fi, err := os.Stat(destinationPath)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(transformedSize).Should(Equal(fi.Size()))
		})
	})

	Context("when the file is a .tar.gz", func() {
		BeforeEach(func() {
			sourcePath = filepath.Join(scratch, "file.tar.gz")

			test_helper.CreateTarGZArchive(sourcePath, archiveFiles)
		})

		It("does not error", func() {
			Ω(transformErr).ShouldNot(HaveOccurred())
		})

		It("gzip uncompresses it to a .tar", func() {
			verifyTarFile(destinationPath)
		})

		It("deletes the original file", func() {
			_, err := os.Stat(sourcePath)
			Ω(err).Should(HaveOccurred())
		})

		It("returns the correct number of bytes written", func() {
			fi, err := os.Stat(destinationPath)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(fi.Size()).Should(Equal(transformedSize))
		})
	})

	Context("when the file is a .zip", func() {
		BeforeEach(func() {
			sourcePath = filepath.Join(scratch, "file.zip")

			test_helper.CreateZipArchive(sourcePath, archiveFiles)
		})

		It("does not error", func() {
			Ω(transformErr).ShouldNot(HaveOccurred())
		})

		It("transforms it to a .tar", func() {
			verifyTarFile(destinationPath)
		})

		It("deletes the original file", func() {
			_, err := os.Stat(sourcePath)
			Ω(err).Should(HaveOccurred())
		})

		It("returns the correct number of bytes written", func() {
			fi, err := os.Stat(destinationPath)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(fi.Size()).Should(Equal(transformedSize))
		})
	})

	Context("when the file is a .mp3", func() {
		BeforeEach(func() {
			sourcePath = filepath.Join(scratch, "bogus")

			err := ioutil.WriteFile(sourcePath, []byte("bogus"), 0755)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("blows up horribly", func() {
			Ω(transformErr).Should(Equal(ErrUnknownArchiveFormat))
		})
	})
})
