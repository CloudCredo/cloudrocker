// +build !linux,!freebsd linux,!amd64 freebsd,!cgo

package mount

func mount(device, target, mType string, flag uintptr, data string) error {
	panic("Not implemented")
}

func unmount(target string, flag int) error {
	panic("Not implemented")
}
