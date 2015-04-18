package cacheddownloader_test

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	Url "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega/ghttp"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/pivotal-golang/cacheddownloader"
)

const MAX_CONCURRENT_DOWNLOADS = 10

func computeMd5(key string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(key)))
}

var _ = Describe("File cache", func() {
	var (
		cache           CachedDownloader
		cachedPath      string
		uncachedPath    string
		maxSizeInBytes  int64
		cacheKey        string
		downloadContent []byte
		url             *Url.URL
		server          *ghttp.Server
		cancelChan      chan struct{}
	)

	BeforeEach(func() {
		var err error
		cachedPath, err = ioutil.TempDir("", "test_file_cached")
		Ω(err).ShouldNot(HaveOccurred())

		uncachedPath, err = ioutil.TempDir("", "test_file_uncached")
		Ω(err).ShouldNot(HaveOccurred())

		maxSizeInBytes = 1024

		cacheKey = "the-cache-key"

		cache = New(cachedPath, uncachedPath, maxSizeInBytes, 1*time.Second, MAX_CONCURRENT_DOWNLOADS, false)
		server = ghttp.NewServer()

		url, err = Url.Parse(server.URL() + "/my_file")
		Ω(err).ShouldNot(HaveOccurred())

		cancelChan = make(chan struct{})
	})

	AfterEach(func() {
		os.RemoveAll(cachedPath)
		os.RemoveAll(uncachedPath)
	})

	var (
		file     io.ReadCloser
		fileSize int64
		err      error
	)

	Describe("when the cache folder does not exist", func() {
		It("should create it", func() {
			os.RemoveAll(cachedPath)
			cache = New(cachedPath, uncachedPath, maxSizeInBytes, time.Second, MAX_CONCURRENT_DOWNLOADS, false)
			_, err := os.Stat(cachedPath)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("when the cache folder has stuff in it", func() {
		It("should nuke that stuff", func() {
			filename := filepath.Join(cachedPath, "last_nights_dinner")
			ioutil.WriteFile(filename, []byte("leftovers"), 0666)
			cache = New(cachedPath, uncachedPath, maxSizeInBytes, time.Second, MAX_CONCURRENT_DOWNLOADS, false)
			_, err := os.Stat(filename)
			Ω(err).Should(HaveOccurred())
		})
	})

	Describe("When providing a file that should not be cached", func() {
		Context("when the download succeeds", func() {
			BeforeEach(func() {
				downloadContent = []byte("777")

				header := http.Header{}
				header.Set("ETag", "foo")
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/my_file"),
					ghttp.RespondWith(http.StatusOK, string(downloadContent), header),
				))

				file, fileSize, err = cache.Fetch(url, "", NoopTransform, cancelChan)
			})

			It("should not error", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("should return a readCloser that streams the file", func() {
				Ω(file).ShouldNot(BeNil())
				Ω(fileSize).Should(BeNumerically("==", 3))
				Ω(ioutil.ReadAll(file)).Should(Equal(downloadContent))
			})

			It("should delete the file when we close the readCloser", func() {
				err := file.Close()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(ioutil.ReadDir(uncachedPath)).Should(HaveLen(0))
				Ω(ioutil.ReadDir(cachedPath)).Should(HaveLen(0))
			})
		})

		Context("when the download fails", func() {
			BeforeEach(func() {
				server.AllowUnhandledRequests = true //will 500 for any attempted requests
				file, fileSize, err = cache.Fetch(url, "", NoopTransform, cancelChan)
			})

			It("should return an error and no file", func() {
				Ω(file).Should(BeNil())
				Ω(fileSize).Should(BeNumerically("==", 0))
				Ω(err).Should(HaveOccurred())
			})

			It("should clean up after itself", func() {
				Ω(ioutil.ReadDir(uncachedPath)).Should(HaveLen(0))
				Ω(ioutil.ReadDir(cachedPath)).Should(HaveLen(0))
			})
		})
	})

	Describe("When providing a file that should be cached", func() {
		var cacheFilePath string
		var returnedHeader http.Header

		BeforeEach(func() {
			cacheFilePath = filepath.Join(cachedPath, computeMd5(cacheKey))
			returnedHeader = http.Header{}
			returnedHeader.Set("ETag", "my-original-etag")
		})

		Context("when the file is not in the cache", func() {
			var (
				transformer CacheTransformer

				fetchedFile     io.ReadCloser
				fetchedFileSize int64
				fetchErr        error
			)

			BeforeEach(func() {
				transformer = NoopTransform
			})

			JustBeforeEach(func() {
				fetchedFile, fetchedFileSize, fetchErr = cache.Fetch(url, cacheKey, transformer, cancelChan)
			})

			Context("when the download succeeds", func() {
				BeforeEach(func() {
					downloadContent = []byte(strings.Repeat("7", int(maxSizeInBytes/2)))
					server.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/my_file"),
						http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
							Ω(req.Header.Get("If-None-Match")).Should(BeEmpty())
						}),
						ghttp.RespondWith(http.StatusOK, string(downloadContent), returnedHeader),
					))
				})

				It("should not error", func() {
					Ω(fetchErr).ShouldNot(HaveOccurred())
				})

				It("should return a readCloser that streams the file", func() {
					Ω(fetchedFile).ShouldNot(BeNil())
					Ω(fetchedFileSize).Should(BeNumerically("==", maxSizeInBytes/2))
					Ω(ioutil.ReadAll(fetchedFile)).Should(Equal(downloadContent))
				})

				It("should return a file within the cache", func() {
					Ω(ioutil.ReadDir(cachedPath)).Should(HaveLen(1))
				})

				It("should remove any temporary assets generated along the way", func() {
					Ω(ioutil.ReadDir(uncachedPath)).Should(HaveLen(0))
				})

				Describe("downloading with a transformer", func() {
					BeforeEach(func() {
						transformer = func(source string, destination string) (int64, error) {
							err := ioutil.WriteFile(destination, []byte("hello tmp"), 0644)
							Ω(err).ShouldNot(HaveOccurred())

							return 100, err
						}
					})

					It("passes the download through the transformer", func() {
						Ω(fetchErr).ShouldNot(HaveOccurred())

						content, err := ioutil.ReadAll(fetchedFile)
						Ω(err).ShouldNot(HaveOccurred())
						Ω(string(content)).Should(Equal("hello tmp"))
					})
				})
			})

			Context("when the download succeeds but does not have an ETag", func() {
				BeforeEach(func() {
					downloadContent = []byte(strings.Repeat("7", int(maxSizeInBytes/2)))
					server.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/my_file"),
						http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
							Ω(req.Header.Get("If-None-Match")).Should(BeEmpty())
						}),
						ghttp.RespondWith(http.StatusOK, string(downloadContent)),
					))
				})

				It("should not error", func() {
					Ω(fetchErr).ShouldNot(HaveOccurred())
				})

				It("should return a readCloser that streams the file", func() {
					Ω(fetchedFile).ShouldNot(BeNil())
					Ω(ioutil.ReadAll(fetchedFile)).Should(Equal(downloadContent))
				})

				It("should not store the file", func() {
					fetchedFile.Close()
					Ω(ioutil.ReadDir(cachedPath)).Should(HaveLen(0))
					Ω(ioutil.ReadDir(uncachedPath)).Should(HaveLen(0))
				})
			})

			Context("when the download fails", func() {
				BeforeEach(func() {
					server.AllowUnhandledRequests = true //will 500 for any attempted requests
				})

				It("should return an error and no file", func() {
					Ω(fetchedFile).Should(BeNil())
					Ω(fetchErr).Should(HaveOccurred())
				})

				It("should clean up after itself", func() {
					Ω(ioutil.ReadDir(cachedPath)).Should(HaveLen(0))
					Ω(ioutil.ReadDir(uncachedPath)).Should(HaveLen(0))
				})
			})
		})

		Context("when the file is already on disk in the cache", func() {
			var fileContent []byte
			var status int
			var downloadContent string

			BeforeEach(func() {
				status = http.StatusOK
				fileContent = []byte("now you see it")

				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/my_file"),
					ghttp.RespondWith(http.StatusOK, string(fileContent), returnedHeader),
				))

				f, s, _ := cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
				defer f.Close()
				Ω(s).Should(BeNumerically("==", len(fileContent)))

				downloadContent = "now you don't"

				etag := "second-round-etag"
				returnedHeader.Set("ETag", etag)

				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/my_file"),
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						Ω(req.Header.Get("If-None-Match")).Should(Equal("my-original-etag"))
					}),
					ghttp.RespondWithPtr(&status, &downloadContent, returnedHeader),
				))
			})

			It("should perform the request with the correct modified headers", func() {
				cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
				Ω(server.ReceivedRequests()).Should(HaveLen(2))
			})

			Context("if the file has been modified", func() {
				BeforeEach(func() {
					status = http.StatusOK
				})

				It("should redownload the file", func() {
					f, _, _ := cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
					defer f.Close()

					paths, _ := filepath.Glob(cacheFilePath + "*")
					Ω(ioutil.ReadFile(paths[0])).Should(Equal([]byte(downloadContent)))
				})

				It("should return a readcloser pointing to the file", func() {
					file, fileSize, err := cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(ioutil.ReadAll(file)).Should(Equal([]byte(downloadContent)))
					Ω(fileSize).Should(BeNumerically("==", len(downloadContent)))
				})

				It("should have put the file in the cache", func() {
					f, s, err := cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
					f.Close()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(ioutil.ReadDir(cachedPath)).Should(HaveLen(1))
					Ω(ioutil.ReadDir(uncachedPath)).Should(HaveLen(0))
					Ω(s).Should(BeNumerically("==", len(downloadContent)))
				})
			})

			Context("if the file has been modified, but the new file has no etag", func() {
				BeforeEach(func() {
					status = http.StatusOK
					returnedHeader.Del("ETag")
				})

				It("should return a readcloser pointing to the file", func() {
					file, fileSize, err := cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(ioutil.ReadAll(file)).Should(Equal([]byte(downloadContent)))
					Ω(fileSize).Should(BeNumerically("==", len(downloadContent)))
				})

				It("should have removed the file from the cache", func() {
					f, s, err := cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
					f.Close()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(ioutil.ReadDir(cachedPath)).Should(HaveLen(0))
					Ω(ioutil.ReadDir(uncachedPath)).Should(HaveLen(0))
					Ω(s).Should(BeNumerically("==", len(downloadContent)))
				})
			})

			Context("if the file has not been modified", func() {
				BeforeEach(func() {
					status = http.StatusNotModified
				})

				It("should not redownload the file", func() {
					f, s, err := cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
					Ω(err).ShouldNot(HaveOccurred())
					defer f.Close()

					paths, _ := filepath.Glob(cacheFilePath + "*")
					Ω(ioutil.ReadFile(paths[0])).Should(Equal(fileContent))
					Ω(s).Should(BeZero())
				})

				It("should return a readcloser pointing to the file", func() {
					file, _, err := cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(ioutil.ReadAll(file)).Should(Equal(fileContent))
				})
			})
		})

		Context("when the file size exceeds the total available cache size", func() {
			BeforeEach(func() {
				downloadContent = []byte(strings.Repeat("7", int(maxSizeInBytes*3)))

				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/my_file"),
					ghttp.RespondWith(http.StatusOK, string(downloadContent), returnedHeader),
				))

				file, fileSize, err = cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
			})

			It("should not error", func() {
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("should return a readCloser that streams the file", func() {
				Ω(file).ShouldNot(BeNil())
				Ω(ioutil.ReadAll(file)).Should(Equal(downloadContent))
				Ω(fileSize).Should(BeNumerically("==", maxSizeInBytes*3))
			})

			It("should put the file in the uncached path, then delete it", func() {
				err := file.Close()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(ioutil.ReadDir(cachedPath)).Should(HaveLen(0))
				Ω(ioutil.ReadDir(uncachedPath)).Should(HaveLen(0))
			})

			Context("when the file is downloaded the second time", func() {
				BeforeEach(func() {
					err = file.Close()
					Ω(err).ShouldNot(HaveOccurred())

					server.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/my_file"),
						http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
							Ω(req.Header.Get("If-None-Match")).Should(BeEmpty())
						}),
						ghttp.RespondWith(http.StatusOK, string(downloadContent), returnedHeader),
					))

					file, _, err = cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
				})

				It("should not error", func() {
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("should return a readCloser that streams the file", func() {
					Ω(file).ShouldNot(BeNil())
					Ω(ioutil.ReadAll(file)).Should(Equal(downloadContent))
				})
			})
		})

		Context("when the cache is full", func() {
			fetchFileOfSize := func(name string, size int) {
				downloadContent = []byte(strings.Repeat("7", size))
				url, _ = Url.Parse(server.URL() + "/" + name)
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/"+name),
					ghttp.RespondWith(http.StatusOK, string(downloadContent), returnedHeader),
				))

				cachedFile, _, err := cache.Fetch(url, name, NoopTransform, cancelChan)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(ioutil.ReadAll(cachedFile)).Should(Equal(downloadContent))
				cachedFile.Close()
			}

			BeforeEach(func() {
				fetchFileOfSize("A", int(maxSizeInBytes/4))
				fetchFileOfSize("B", int(maxSizeInBytes/4))
				fetchFileOfSize("C", int(maxSizeInBytes/4))
			})

			It("deletes the oldest cached files until there is space", func() {
				//try to add a file that has size larger
				fetchFileOfSize("D", int(maxSizeInBytes/2)+1)

				Ω(ioutil.ReadDir(cachedPath)).Should(HaveLen(2))

				Ω(filepath.Glob(filepath.Join(cachedPath, computeMd5("A")+"*"))).Should(HaveLen(0))
				Ω(filepath.Glob(filepath.Join(cachedPath, computeMd5("B")+"*"))).Should(HaveLen(0))
				Ω(filepath.Glob(filepath.Join(cachedPath, computeMd5("C")+"*"))).Should(HaveLen(1))
				Ω(filepath.Glob(filepath.Join(cachedPath, computeMd5("D")+"*"))).Should(HaveLen(1))
			})

			Describe("when one of the files has just been read", func() {
				BeforeEach(func() {
					server.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/A"),
						ghttp.RespondWith(http.StatusNotModified, "", returnedHeader),
					))

					url, _ = Url.Parse(server.URL() + "/A")
					// windows timer increments every 15.6ms, without the sleep
					// A, B & C will sometimes have the same timestamp
					time.Sleep(16 * time.Millisecond)
					cache.Fetch(url, "A", NoopTransform, cancelChan)
				})

				It("considers that file to be the newest", func() {
					//try to add a file that has size larger
					fetchFileOfSize("D", int(maxSizeInBytes/2)+1)

					Ω(ioutil.ReadDir(cachedPath)).Should(HaveLen(2))

					Ω(filepath.Glob(filepath.Join(cachedPath, computeMd5("A")+"*"))).Should(HaveLen(1))
					Ω(filepath.Glob(filepath.Join(cachedPath, computeMd5("B")+"*"))).Should(HaveLen(0))
					Ω(filepath.Glob(filepath.Join(cachedPath, computeMd5("C")+"*"))).Should(HaveLen(0))
					Ω(filepath.Glob(filepath.Join(cachedPath, computeMd5("D")+"*"))).Should(HaveLen(1))
				})
			})
		})
	})

	Describe("rate limiting", func() {
		Context("when multiple, concurrent requests occur", func() {
			var barrier chan interface{}
			var results chan bool

			BeforeEach(func() {
				barrier = make(chan interface{}, 1)
				results = make(chan bool, 1)

				downloadContent = []byte(strings.Repeat("7", int(maxSizeInBytes/2)))
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/my_file"),
						http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
							barrier <- nil
							Consistently(results, .5).ShouldNot(Receive())
						}),
						ghttp.RespondWith(http.StatusOK, string(downloadContent)),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/my_file"),
						http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
							results <- true
						}),
						ghttp.RespondWith(http.StatusOK, string(downloadContent)),
					),
				)
			})

			It("processes one at a time", func() {
				go func() {
					cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
					barrier <- nil
				}()
				<-barrier
				cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
				<-barrier
			})
		})

		Context("when cancelling", func() {
			var requestInitiated chan struct{}
			var completeRequest chan struct{}

			BeforeEach(func() {
				requestInitiated = make(chan struct{})
				completeRequest = make(chan struct{})

				server.AllowUnhandledRequests = true
				server.AppendHandlers(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						requestInitiated <- struct{}{}
						<-completeRequest
						w.Write([]byte("response data..."))
					}),
				)

				serverUrl := server.URL() + "/somepath"
				url, _ = url.Parse(serverUrl)
			})

			It("stops waiting", func() {
				errs := make(chan error)

				go func() {
					_, _, err := cache.Fetch(url, cacheKey, NoopTransform, make(chan struct{}))
					errs <- err
				}()
				Eventually(requestInitiated).Should(Receive())

				close(cancelChan)

				_, _, err := cache.Fetch(url, cacheKey, NoopTransform, cancelChan)
				Ω(err).Should(Equal(ErrDownloadCancelled))

				close(completeRequest)
				Eventually(errs).Should(Receive(BeNil()))
			})
		})
	})
})

type constTransformer struct {
	file string
	size int64
	err  error
}

func (t constTransformer) ConstTransform(path string) (string, int64, error) {
	return t.file, t.size, t.err
}
