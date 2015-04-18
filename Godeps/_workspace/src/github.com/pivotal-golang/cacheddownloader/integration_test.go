package cacheddownloader_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"time"

	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudcredo/cloudrocker/Godeps/_workspace/src/github.com/pivotal-golang/cacheddownloader"
)

var _ = Describe("Integration", func() {
	var (
		server              *httptest.Server
		serverPath          string
		cachedPath          string
		uncachedPath        string
		cacheMaxSizeInBytes int64         = 1024
		downloadTimeout     time.Duration = time.Second
		downloader          cacheddownloader.CachedDownloader
		url                 *url.URL
	)

	BeforeEach(func() {
		var err error

		serverPath, err = ioutil.TempDir("", "cached_downloader_integration_server")
		Ω(err).ShouldNot(HaveOccurred())

		cachedPath, err = ioutil.TempDir("", "cached_downloader_integration_cache")
		Ω(err).ShouldNot(HaveOccurred())

		uncachedPath, err = ioutil.TempDir("", "cached_downloader_integration_uncached")
		Ω(err).ShouldNot(HaveOccurred())

		handler := http.FileServer(http.Dir(serverPath))
		server = httptest.NewServer(handler)

		url, err = url.Parse(server.URL + "/file")
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(serverPath)
		os.RemoveAll(cachedPath)
		os.RemoveAll(uncachedPath)
		server.Close()
	})

	fetch := func() ([]byte, time.Time) {
		url, err := url.Parse(server.URL + "/file")
		Ω(err).ShouldNot(HaveOccurred())

		reader, _, err := downloader.Fetch(url, "the-cache-key", cacheddownloader.NoopTransform, make(chan struct{}))
		Ω(err).ShouldNot(HaveOccurred())
		defer reader.Close()

		readData, err := ioutil.ReadAll(reader)
		Ω(err).ShouldNot(HaveOccurred())

		cacheContents, err := ioutil.ReadDir(cachedPath)
		Ω(cacheContents).Should(HaveLen(1))
		Ω(err).ShouldNot(HaveOccurred())

		content, err := ioutil.ReadFile(filepath.Join(cachedPath, cacheContents[0].Name()))
		Ω(err).ShouldNot(HaveOccurred())

		Ω(readData).Should(Equal(content))

		return content, cacheContents[0].ModTime()
	}

	Describe("Cached Downloader", func() {
		BeforeEach(func() {
			downloader = cacheddownloader.New(cachedPath, uncachedPath, cacheMaxSizeInBytes, downloadTimeout, 10, false)

			// touch a file on disk
			err := ioutil.WriteFile(filepath.Join(serverPath, "file"), []byte("a"), 0666)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("caches downloads", func() {
			// download file once
			content, modTimeBefore := fetch()
			Ω(content).Should(Equal([]byte("a")))

			time.Sleep(time.Second)

			// download again should be cached
			content, modTimeAfter := fetch()
			Ω(content).Should(Equal([]byte("a")))
			Ω(modTimeBefore).Should(Equal(modTimeAfter))

			time.Sleep(time.Second)

			// touch file again
			err := ioutil.WriteFile(filepath.Join(serverPath, "file"), []byte("b"), 0666)
			Ω(err).ShouldNot(HaveOccurred())

			// download again and we should get a file containing "b"
			content, _ = fetch()
			Ω(content).Should(Equal([]byte("b")))
		})
	})
})
