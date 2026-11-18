//go:build windows

package update

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	modKernel32     = syscall.NewLazyDLL("kernel32.dll")
	procMoveFileExW = modKernel32.NewProc("MoveFileExW")
)

const (
	movefileReplaceExisting = 0x1
	movefileWriteThrough    = 0x8
)

// moveFile renames within one volume using MoveFileExW. It needs only DELETE
// access on the source, so a running .exe can be renamed (but not deleted).
func moveFile(from, to string) error {
	fromPtr, err := syscall.UTF16PtrFromString(from)
	if err != nil {
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: err}
	}
	toPtr, err := syscall.UTF16PtrFromString(to)
	if err != nil {
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: err}
	}
	r, _, callErr := procMoveFileExW.Call(
		uintptr(unsafe.Pointer(fromPtr)),
		uintptr(unsafe.Pointer(toPtr)),
		movefileReplaceExisting|movefileWriteThrough,
	)
	if r == 0 {
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: callErr}
	}
	return nil
}
