//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	modadvapi32             = syscall.NewLazyDLL("advapi32.dll")
	modshell32              = syscall.NewLazyDLL("shell32.dll")
	modkernel32             = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcessToken    = modadvapi32.NewProc("OpenProcessToken")
	procGetTokenInformation = modadvapi32.NewProc("GetTokenInformation")
	procShellExecuteW       = modshell32.NewProc("ShellExecuteW")
	procGetCurrentProcess   = modkernel32.NewProc("GetCurrentProcess")
)

const (
	tokenQuery     = 0x0008
	tokenElevation = 20 // TokenElevation class
	swShowNormal   = 1
)

// isElevated reports whether the current process runs with an elevated token.
func isElevated() bool {
	cur, _, _ := procGetCurrentProcess.Call()
	var token syscall.Handle
	ret, _, _ := procOpenProcessToken.Call(cur, tokenQuery, uintptr(unsafe.Pointer(&token)))
	if ret == 0 {
		return false
	}
	defer syscall.CloseHandle(token)

	var elevation uint32
	var retLen uint32
	ret, _, _ = procGetTokenInformation.Call(
		uintptr(token),
		uintptr(tokenElevation),
		uintptr(unsafe.Pointer(&elevation)),
		unsafe.Sizeof(elevation),
		uintptr(unsafe.Pointer(&retLen)),
	)
	if ret == 0 {
		return false
	}
	return elevation != 0
}

// relaunchElevated starts a new elevated instance of this executable via the
// "runas" verb (triggers the UAC prompt). Returns true if the child started.
func relaunchElevated() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	exe, _ = filepath.Abs(exe)

	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	cwd, _ := syscall.UTF16PtrFromString(filepath.Dir(exe))

	var argPtr *uint16
	if len(os.Args) > 1 {
		argPtr, _ = syscall.UTF16PtrFromString(strings.Join(os.Args[1:], " "))
	}

	ret, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(argPtr)),
		uintptr(unsafe.Pointer(cwd)),
		swShowNormal,
	)
	return ret > 32
}

// ensureElevated relaunches the app elevated when needed. It returns true when
// the current (non-elevated) instance should exit because an elevated child was
// started. If elevation fails or the user cancels the UAC prompt, it returns
// false so the app still opens (tweaks may then fail with access-denied).
func ensureElevated() bool {
	if isElevated() {
		return false
	}
	if relaunchElevated() {
		return true
	}
	return false
}
