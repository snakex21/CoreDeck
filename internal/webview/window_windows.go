//go:build windows

package webview

import (
	"syscall"

	"github.com/webview/webview_go"
)

func maximizeWindow(w webview.WebView, maximize bool) {
	if !maximize {
		return
	}
	w.Dispatch(func() {
		if hwnd := uintptr(w.Window()); hwnd != 0 {
			user32 := syscall.NewLazyDLL("user32.dll")
			user32.NewProc("ShowWindow").Call(hwnd, 3) // SW_MAXIMIZE
		}
	})
}
