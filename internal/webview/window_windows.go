//go:build windows

package webview

import (
	"syscall"
	"unsafe"

	"github.com/webview/webview_go"

	"launcher/internal/windowstate"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procShowWindow    = user32.NewProc("ShowWindow")
	procIsZoomed      = user32.NewProc("IsZoomed")
	procGetWindowRect = user32.NewProc("GetWindowRect")
)

type nativeRect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

func maximizeWindow(w webview.WebView, maximize bool) {
	if !maximize {
		return
	}
	w.Dispatch(func() {
		if hwnd := uintptr(w.Window()); hwnd != 0 {
			procShowWindow.Call(hwnd, 3) // SW_MAXIMIZE
		}
	})
}

func captureWindowState(w webview.WebView, fallback windowstate.State) windowstate.State {
	hwnd := uintptr(w.Window())
	if hwnd == 0 {
		return fallback
	}

	state := fallback
	if zoomed, _, _ := procIsZoomed.Call(hwnd); zoomed != 0 {
		state.Maximized = true
		return state
	}
	state.Maximized = false

	var rect nativeRect
	if ok, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect))); ok == 0 {
		return state
	}
	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)
	if width > 0 && height > 0 {
		state.Width = width
		state.Height = height
	}
	return state
}
