package cacheddownloader_test

import (
	"io"
	"io/ioutil"
	"os"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/pivotal-golang/cacheddownloader"
)

var _ = Describe("FileCache", func() {
	var cache *FileCache
	var cacheDir string
	var err error

	var sourceFile *os.File

	BeforeEach(func() {
		cacheDir, err = ioutil.TempDir("", "cache-test")
		Ω(err).ShouldNot(HaveOccurred())

		cache = NewCache(cacheDir, 123424)

		sourceFile = createFile("cache-test-file", "the-file-content")
	})

	AfterEach(func() {
		os.RemoveAll(sourceFile.Name())
		os.RemoveAll(cacheDir)
	})

	Describe("Add", func() {
		var cacheKey string
		var fileSize int64
		var cacheInfo CachingInfoType
		var readCloser io.ReadCloser

		BeforeEach(func() {
			cacheKey = "the-cache-key"
			fileSize = 100
			cacheInfo = CachingInfoType{}
		})

		It("fails if room cannot be allocated", func() {
			var err error
			readCloser, err = cache.Add(cacheKey, sourceFile.Name(), 250000, cacheInfo)
			Ω(err).Should(Equal(NotEnoughSpace))
			Ω(readCloser).Should(BeNil())
		})

		Context("when closed is called", func() {
			JustBeforeEach(func() {
				var err error
				readCloser, err = cache.Add(cacheKey, sourceFile.Name(), fileSize, cacheInfo)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(readCloser).ShouldNot(BeNil())
			})

			Context("once", func() {
				It("succeeds and has 1 file in the cache", func() {
					Ω(readCloser.Close()).ShouldNot(HaveOccurred())
					Ω(filenamesInDir(cacheDir)).Should(HaveLen(1))
				})
			})

			Context("more than once", func() {
				It("fails", func() {
					Ω(readCloser.Close()).ShouldNot(HaveOccurred())
					Ω(readCloser.Close()).Should(HaveOccurred())
				})
			})
		})

		Context("when the cache is empty", func() {
			JustBeforeEach(func() {
				var err error
				readCloser, err = cache.Add(cacheKey, sourceFile.Name(), fileSize, cacheInfo)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(readCloser).ShouldNot(BeNil())
			})

			AfterEach(func() {
				readCloser.Close()
			})

			It("returns a reader", func() {
				content, err := ioutil.ReadAll(readCloser)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(content)).Should(Equal("the-file-content"))
			})

			It("has 1 file in the cache", func() {
				Ω(filenamesInDir(cacheDir)).Should(HaveLen(1))
			})
		})

		Context("when a cachekey exists", func() {
			var newSourceFile *os.File
			var newFileSize int64
			var newCacheInfo CachingInfoType
			var newReader io.ReadCloser

			BeforeEach(func() {
				newSourceFile = createFile("cache-test-file", "new-file-content")
				newFileSize = fileSize
				newCacheInfo = cacheInfo
			})

			JustBeforeEach(func() {
				readCloser, err = cache.Add(cacheKey, sourceFile.Name(), fileSize, cacheInfo)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(readCloser).ShouldNot(BeNil())
			})

			AfterEach(func() {
				readCloser.Close()
				os.RemoveAll(newSourceFile.Name())
			})

			Context("when adding the same cache key with identical info", func() {
				It("ignores the add", func() {
					reader, err := cache.Add(cacheKey, newSourceFile.Name(), fileSize, cacheInfo)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(reader).ShouldNot(BeNil())
				})
			})

			Context("when a adding the same cache key and different info", func() {
				JustBeforeEach(func() {
					var err error
					newReader, err = cache.Add(cacheKey, newSourceFile.Name(), newFileSize, newCacheInfo)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(newReader).ShouldNot(BeNil())
				})

				AfterEach(func() {
					newReader.Close()
				})

				Context("different file size", func() {
					BeforeEach(func() {
						newFileSize = fileSize - 1
					})

					It("returns a reader for the new content", func() {
						content, err := ioutil.ReadAll(newReader)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(string(content)).Should(Equal("new-file-content"))
					})

					It("has files in the cache", func() {
						Ω(filenamesInDir(cacheDir)).Should(HaveLen(2))
						Ω(readCloser.Close()).ShouldNot(HaveOccurred())
						Ω(filenamesInDir(cacheDir)).Should(HaveLen(1))
					})

				})

				Context("different caching info", func() {
					BeforeEach(func() {
						newCacheInfo = CachingInfoType{
							LastModified: "1234",
						}
					})

					It("returns a reader for the new content", func() {
						content, err := ioutil.ReadAll(newReader)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(string(content)).Should(Equal("new-file-content"))
					})

					It("has files in the cache", func() {
						Ω(filenamesInDir(cacheDir)).Should(HaveLen(2))
						Ω(readCloser.Close()).ShouldNot(HaveOccurred())
						Ω(filenamesInDir(cacheDir)).Should(HaveLen(1))
					})

					It("still allows the previous reader to read", func() {
						content, err := ioutil.ReadAll(readCloser)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(string(content)).Should(Equal("the-file-content"))
					})
				})
			})
		})
	})

	Describe("Get", func() {
		var cacheKey string
		var fileSize int64
		var cacheInfo CachingInfoType

		BeforeEach(func() {
			cacheKey = "key"
			fileSize = 100
			cacheInfo = CachingInfoType{}
		})

		Context("when there is nothing", func() {
			It("returns nothing", func() {
				reader, ci, err := cache.Get(cacheKey)
				Ω(err).Should(Equal(EntryNotFound))
				Ω(reader).Should(BeNil())
				Ω(ci).Should(Equal(cacheInfo))
			})
		})

		Context("when there is an item", func() {
			BeforeEach(func() {
				cacheInfo.LastModified = "1234"
				reader, err := cache.Add(cacheKey, sourceFile.Name(), fileSize, cacheInfo)
				Ω(err).ShouldNot(HaveOccurred())
				reader.Close()
			})

			It("returns a reader for the item", func() {
				reader, ci, err := cache.Get(cacheKey)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(reader).ShouldNot(BeNil())
				Ω(ci).Should(Equal(cacheInfo))

				content, err := ioutil.ReadAll(reader)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(content)).Should(Equal("the-file-content"))
			})

			Context("when the item is replaced", func() {
				var newSourceFile *os.File

				JustBeforeEach(func() {
					newSourceFile = createFile("cache-test-file", "new-file-content")

					cacheInfo.LastModified = "123"
					reader, err := cache.Add(cacheKey, newSourceFile.Name(), fileSize, cacheInfo)
					Ω(err).ShouldNot(HaveOccurred())
					reader.Close()
				})

				AfterEach(func() {
					os.RemoveAll(newSourceFile.Name())
				})

				It("gets the new item", func() {
					reader, ci, err := cache.Get(cacheKey)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(reader).ShouldNot(BeNil())
					Ω(ci).Should(Equal(cacheInfo))

					content, err := ioutil.ReadAll(reader)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(string(content)).Should(Equal("new-file-content"))
				})

				Context("when a get is issued before a replace", func() {
					var reader io.ReadCloser
					BeforeEach(func() {
						var err error
						reader, _, err = cache.Get(cacheKey)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(reader).ShouldNot(BeNil())

						Ω(filenamesInDir(cacheDir)).Should(HaveLen(1))
					})

					It("the old file is removed when closed", func() {
						reader.Close()
						Ω(filenamesInDir(cacheDir)).Should(HaveLen(1))
					})
				})
			})
		})
	})

	Describe("Remove", func() {
		var cacheKey string
		var cacheInfo CachingInfoType

		BeforeEach(func() {
			cacheKey = "key"

			cacheInfo.LastModified = "1234"
			reader, err := cache.Add(cacheKey, sourceFile.Name(), 100, cacheInfo)
			Ω(err).ShouldNot(HaveOccurred())
			reader.Close()
		})

		Context("when the key does not exist", func() {
			It("does not fail", func() {
				Ω(func() { cache.Remove("bogus") }).ShouldNot(Panic())
			})
		})

		Context("when the key exists", func() {
			It("removes the file in the cache", func() {
				Ω(filenamesInDir(cacheDir)).Should(HaveLen(1))
				cache.Remove(cacheKey)
				Ω(filenamesInDir(cacheDir)).Should(HaveLen(0))
			})
		})

		Context("when a get is issued first", func() {
			It("removes the file after a close", func() {
				reader, _, err := cache.Get(cacheKey)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(reader).ShouldNot(BeNil())
				Ω(filenamesInDir(cacheDir)).Should(HaveLen(1))

				cache.Remove(cacheKey)
				Ω(filenamesInDir(cacheDir)).Should(HaveLen(1))

				reader.Close()
				Ω(filenamesInDir(cacheDir)).Should(HaveLen(0))
			})
		})
	})
})

func createFile(filename string, content string) *os.File {
	sourceFile, err := ioutil.TempFile("", filename)
	Ω(err).ShouldNot(HaveOccurred())
	sourceFile.WriteString(content)
	sourceFile.Close()

	return sourceFile
}

func filenamesInDir(dir string) []string {
	entries, err := ioutil.ReadDir(dir)
	Ω(err).ShouldNot(HaveOccurred())

	result := []string{}
	for _, entry := range entries {
		result = append(result, entry.Name())
	}

	return result
}
