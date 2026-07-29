//go:build !windows

package webview

import "github.com/webview/webview_go"

func setNativeWindowIconDelayed(_ webview.WebView, _ string) {}
