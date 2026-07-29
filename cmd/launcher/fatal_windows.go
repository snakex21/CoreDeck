//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

func showFatalError(message string) {
	title, _ := syscall.UTF16PtrFromString("CoreDeck — błąd")
	text, _ := syscall.UTF16PtrFromString(message)
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBoxW := user32.NewProc("MessageBoxW")
	_, _, _ = messageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(text)),
		uintptr(unsafe.Pointer(title)),
		0x00000010,
	)
}
