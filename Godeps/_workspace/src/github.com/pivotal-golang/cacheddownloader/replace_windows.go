package cacheddownloader

import (
	"syscall"
	"unsafe"
)

// Replaces `dst' with `src' atomically. Under linux we only have to
// call os.Rename(), on windows os.Rename() will error if the
// destination exists already. The replace function serves as a
// unified interface on both platforms.
func replace(src, dst string) error {
	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		return err
	}
	defer syscall.FreeLibrary(kernel32)
	moveFileExUnicode, err := syscall.GetProcAddress(kernel32, "MoveFileExW")
	if err != nil {
		return err
	}

	srcString, err := syscall.UTF16PtrFromString(src)
	if err != nil {
		return err
	}

	dstString, err := syscall.UTF16PtrFromString(dst)
	if err != nil {
		return err
	}

	srcPtr := uintptr(unsafe.Pointer(srcString))
	dstPtr := uintptr(unsafe.Pointer(dstString))

	MOVEFILE_REPLACE_EXISTING := 0x1
	flag := uintptr(MOVEFILE_REPLACE_EXISTING)

	_, _, callErr := syscall.Syscall(uintptr(moveFileExUnicode), 3, srcPtr, dstPtr, flag)
	if callErr != 0 {
		return callErr
	}

	return nil
}
