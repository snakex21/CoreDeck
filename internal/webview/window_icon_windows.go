//go:build windows

package webview

import (
	"log/slog"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/webview/webview_go"
)

var (
	iconUser32               = syscall.NewLazyDLL("user32.dll")
	iconProcLoadImageW       = iconUser32.NewProc("LoadImageW")
	iconProcSendMessageW     = iconUser32.NewProc("SendMessageW")
	iconProcSetClassLongPtrW = iconUser32.NewProc("SetClassLongPtrW")
	iconProcGetAncestor      = iconUser32.NewProc("GetAncestor")
	iconProcSetWindowPos     = iconUser32.NewProc("SetWindowPos")
)

const (
	wmSetIcon      = 0x0080
	iconSmall      = 0
	iconBig        = 1
	imageIcon      = 1
	lrLoadFromFile = 0x0010
	gaRoot         = 2

	swpNoSize       = 0x0001
	swpNoMove       = 0x0002
	swpNoZOrder     = 0x0004
	swpFrameChanged = 0x0020
)

var (
	gclpHIcon   = ^uintptr(13) // -14
	gclpHIconSm = ^uintptr(33) // -34
)

func setNativeWindowIconDelayed(w webview.WebView, iconPath string) {
	if iconPath == "" {
		return
	}
	go func() {
		// WebView2 may finish native window initialization after Navigate().
		time.Sleep(750 * time.Millisecond)
		w.Dispatch(func() {
			setNativeWindowIcon(w.Window(), iconPath)
		})
	}()
}

func setNativeWindowIcon(hwnd unsafe.Pointer, iconPath string) {
	if hwnd == nil || iconPath == "" {
		return
	}
	if _, err := os.Stat(iconPath); err != nil {
		slog.Warn("window icon file not found", "path", iconPath, "error", err)
		return
	}

	pathPtr, err := syscall.UTF16PtrFromString(iconPath)
	if err != nil {
		slog.Warn("window icon path conversion failed", "error", err)
		return
	}

	hIcon, _, _ := iconProcLoadImageW.Call(
		0,
		uintptr(unsafe.Pointer(pathPtr)),
		imageIcon,
		0,
		0,
		lrLoadFromFile,
	)
	if hIcon == 0 {
		slog.Warn("window icon load failed", "path", iconPath)
		return
	}

	setIconOnWindow(uintptr(hwnd), hIcon)
	if root, _, _ := iconProcGetAncestor.Call(uintptr(hwnd), gaRoot); root != 0 && root != uintptr(hwnd) {
		setIconOnWindow(root, hIcon)
	}
}

func setIconOnWindow(hwnd uintptr, hIcon uintptr) {
	iconProcSendMessageW.Call(hwnd, wmSetIcon, iconBig, hIcon)
	iconProcSendMessageW.Call(hwnd, wmSetIcon, iconSmall, hIcon)
	iconProcSetClassLongPtrW.Call(hwnd, gclpHIcon, hIcon)
	iconProcSetClassLongPtrW.Call(hwnd, gclpHIconSm, hIcon)
	iconProcSetWindowPos.Call(hwnd, 0, 0, 0, 0, 0, swpNoMove|swpNoSize|swpNoZOrder|swpFrameChanged)
}
