// +build !windows

package cacheddownloader

import "os"

// if you are wondering why we have this function, see `replace'
// implementation in replace_windows.go
func replace(src, dst string) error {
	return os.Rename(src, dst)
}
