# Local copy of github.com/webview/webview_go

Upstream: github.com/webview/webview_go v0.0.0-20240831120633-6173450d4dd6 (MIT,
see LICENSE; `libs/mswebview2` is the Microsoft WebView2 SDK header, BSD-3).
Used through `replace github.com/webview/webview_go => ./third_party/webview_go`
in CoreDeck's `go.mod`. The same patched copy is used by the author's other
Go + WebView2 apps (WebDownloader, BenchForge, Game Launcher).

Changes against upstream:

1. `libs/webview/include/webview.h` (`embed()`): when the environment
   variable `WEBVIEW_GO_USER_DATA_FOLDER` is set, it replaces the hard-coded
   WebView2 profile folder `%APPDATA%\<exe name>`. CoreDeck sets it to
   `data\webview` next to `CoreDeck.exe`, so the profile (theme, language,
   local storage) travels with the portable drive. The official
   `WEBVIEW2_USER_DATA_FOLDER` variable does not work here because
   webview_go's built-in loader bypasses WebView2Loader.dll, which is what
   reads it.

2. `webview.go` (`Dispatch()`): upstream read the callback slot (`index`)
   after releasing the mutex, so two goroutines dispatching at the same time
   could schedule the same or an empty slot and the UI thread then called a
   nil func and crashed. The slot is captured under the lock and nil
   callbacks are ignored.
