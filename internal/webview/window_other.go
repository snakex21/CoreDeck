//go:build !windows

package webview

import "github.com/webview/webview_go"

func maximizeWindow(w webview.WebView, maximize bool) {}
