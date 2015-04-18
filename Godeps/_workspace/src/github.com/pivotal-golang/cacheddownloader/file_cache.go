package cacheddownloader

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	lock           = &sync.Mutex{}
	EntryNotFound  = errors.New("Entry Not Found")
	NotEnoughSpace = errors.New("No space available")
)

type FileCache struct {
	cachedPath     string
	maxSizeInBytes int64
	entries        map[string]*fileCacheEntry
	cacheFilePaths map[string]string
	seq            uint64
}

type fileCacheEntry struct {
	size        int64
	access      time.Time
	cachingInfo CachingInfoType
	filePath    string
	inuseCount  int
}

func NewCache(dir string, maxSizeInBytes int64) *FileCache {
	return &FileCache{
		cachedPath:     dir,
		maxSizeInBytes: maxSizeInBytes,
		entries:        map[string]*fileCacheEntry{},
		cacheFilePaths: map[string]string{},
		seq:            0,
	}
}

func newFileCacheEntry(cachePath string, size int64, cachingInfo CachingInfoType) *fileCacheEntry {
	return &fileCacheEntry{
		size:        size,
		filePath:    cachePath,
		access:      time.Now(),
		cachingInfo: cachingInfo,
		inuseCount:  1,
	}
}

func (e *fileCacheEntry) incrementUse() {
	e.inuseCount++
}

func (e *fileCacheEntry) decrementUse() {
	e.inuseCount--
	count := e.inuseCount

	if count == 0 {
		os.RemoveAll(e.filePath)
	}
}

func (e *fileCacheEntry) readCloser() (*CachedFile, error) {
	f, err := os.Open(e.filePath)
	if err != nil {
		return nil, err
	}

	e.incrementUse()
	readCloser := NewFileCloser(f, func(filePath string) {
		lock.Lock()
		e.decrementUse()
		lock.Unlock()
	})

	return readCloser, nil
}

func (c *FileCache) Add(cacheKey, sourcePath string, size int64, cachingInfo CachingInfoType) (*CachedFile, error) {
	lock.Lock()
	defer lock.Unlock()

	oldEntry := c.entries[cacheKey]
	if oldEntry != nil {
		if size == oldEntry.size && cachingInfo.Equal(oldEntry.cachingInfo) {
			err := os.Remove(sourcePath)
			if err != nil {
				return nil, err
			}

			return oldEntry.readCloser()
		}
	}

	if !c.makeRoom(size) {
		//file does not fit in cache...
		return nil, NotEnoughSpace
	}

	c.seq++
	uniqueName := fmt.Sprintf("%s-%d-%d", cacheKey, time.Now().UnixNano(), c.seq)
	cachePath := filepath.Join(c.cachedPath, uniqueName)

	err := replace(sourcePath, cachePath)
	if err != nil {
		return nil, err
	}

	newEntry := newFileCacheEntry(cachePath, size, cachingInfo)
	c.entries[cacheKey] = newEntry
	if oldEntry != nil {
		oldEntry.decrementUse()
	}
	return newEntry.readCloser()
}

func (c *FileCache) Get(cacheKey string) (*CachedFile, CachingInfoType, error) {
	lock.Lock()
	defer lock.Unlock()

	entry := c.entries[cacheKey]
	if entry == nil {
		return nil, CachingInfoType{}, EntryNotFound
	}

	entry.access = time.Now()
	readCloser, err := entry.readCloser()
	if err != nil {
		return nil, CachingInfoType{}, err
	}

	return readCloser, entry.cachingInfo, nil
}

func (c *FileCache) Remove(cacheKey string) {
	lock.Lock()
	c.remove(cacheKey)
	lock.Unlock()
}

func (c *FileCache) remove(cacheKey string) {
	entry := c.entries[cacheKey]
	if entry != nil {
		entry.decrementUse()
		delete(c.entries, cacheKey)
	}
}

func (c *FileCache) makeRoom(size int64) bool {
	if size > c.maxSizeInBytes {
		return false
	}

	usedSpace := c.usedSpace()
	for c.maxSizeInBytes < usedSpace+size {
		var oldestEntry *fileCacheEntry
		oldestAccessTime, oldestCacheKey := time.Now(), ""
		for ck, f := range c.entries {
			if f.access.Before(oldestAccessTime) {
				oldestAccessTime = f.access
				oldestEntry = f
				oldestCacheKey = ck
			}
		}

		usedSpace -= oldestEntry.size
		c.remove(oldestCacheKey)
	}

	return true
}

func (c *FileCache) usedSpace() int64 {
	space := int64(0)
	for _, f := range c.entries {
		space += f.size
	}
	return space
}
