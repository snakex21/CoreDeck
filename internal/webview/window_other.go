//go:build !windows

package webview

import "github.com/webview/webview_go"

import "launcher/internal/windowstate"

func maximizeWindow(w webview.WebView, maximize bool) {}

func captureWindowState(w webview.WebView, fallback windowstate.State) windowstate.State {
	return fallback
}
