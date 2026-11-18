//go:build windows

package programs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	modShell32         = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW  = modShell32.NewProc("ShellExecuteW")
	errShellExecFailed = syscall.Errno(0)
)

const swShowNormal = 1

// startDetached starts a program the way Explorer does on double-click
// (ShellExecute "open"): manifests asking for administrator rights, the UAC
// prompt and compatibility shims are handled by Windows, and the program
// keeps running after CoreDeck closes. Command-line tools get a console
// window that stays open, so their output can be read.
func startDetached(path string) error {
	dir := filepath.Dir(path)
	if isConsoleExecutable(path) {
		comspec := strings.TrimSpace(os.Getenv("ComSpec"))
		if comspec == "" {
			comspec = "cmd.exe"
		}
		return shellExecute(comspec, `/k "`+path+`"`, dir)
	}
	return shellExecute(path, "", dir)
}

func shellExecute(file, args, dir string) error {
	verb, _ := syscall.UTF16PtrFromString("open")
	filePtr, err := syscall.UTF16PtrFromString(file)
	if err != nil {
		return err
	}
	dirPtr, _ := syscall.UTF16PtrFromString(dir)
	var argsPtr *uint16
	if args != "" {
		argsPtr, _ = syscall.UTF16PtrFromString(args)
	}
	ret, _, callErr := procShellExecuteW.Call(0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(filePtr)),
		uintptr(unsafe.Pointer(argsPtr)),
		uintptr(unsafe.Pointer(dirPtr)),
		swShowNormal)
	if ret > 32 {
		return nil
	}
	// ShellExecute returns a small error value; GetLastError (callErr) has
	// the real code, e.g. ERROR_CANCELLED when UAC was refused.
	if errno, ok := callErr.(syscall.Errno); ok && errno != errShellExecFailed {
		return errno
	}
	return syscall.Errno(ret)
}

func openInExplorer(folder string) error {
	return exec.Command("explorer", folder).Start()
}
