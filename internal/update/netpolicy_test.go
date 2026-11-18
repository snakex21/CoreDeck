package update

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCheckAssetURL(t *testing.T) {
	prod := Policy{APIURL: ReleasesAPI}
	ok := []string{
		"https://github.com/snakex21/CoreDeck/releases/download/v2.5.0/CoreDeck-Portable.zip",
		"https://github.com/snakex21/CoreDeck/releases/download/v2.5.0-rc.1/SHA256SUMS.txt",
	}
	bad := []string{
		"http://github.com/snakex21/CoreDeck/releases/download/v2.5.0/CoreDeck-Portable.zip",
		"https://github.com/evil/CoreDeck/releases/download/v2.5.0/CoreDeck-Portable.zip",
		"https://github.com.evil.example/snakex21/CoreDeck/releases/download/v2.5.0/x.zip",
		"https://user@github.com/snakex21/CoreDeck/releases/download/v2.5.0/x.zip",
		"https://github.com:8443/snakex21/CoreDeck/releases/download/v2.5.0/x.zip",
		"https://github.com/snakex21/CoreDeck/releases/download/../../../evil/x.zip",
		"https://github.com/snakex21/CoreDeck/releases/download/v1/x.zip?redirect=evil",
		"https://objects.githubusercontent.com/whatever.zip",
		"http://127.0.0.1:8080/dl/CoreDeck-Portable.zip",
		"",
	}
	for _, u := range ok {
		if err := prod.CheckAssetURL(u); err != nil {
			t.Errorf("%s rejected: %v", u, err)
		}
	}
	for _, u := range bad {
		if err := prod.CheckAssetURL(u); err == nil {
			t.Errorf("%s accepted", u)
		}
	}
	local := Policy{AllowLoopback: true}
	if err := local.CheckAssetURL("http://127.0.0.1:8080/dl/CoreDeck-Portable.zip"); err != nil {
		t.Errorf("loopback rejected with override: %v", err)
	}
	if err := local.CheckAssetURL("http://localhost.evil.example/dl.zip"); err == nil {
		t.Error("non-loopback http accepted with override")
	}
}

func TestCheckRedirectURL(t *testing.T) {
	prod := Policy{}
	for raw, want := range map[string]bool{
		"https://github.com/x":                           true,
		"https://objects.githubusercontent.com/x":        true,
		"https://release-assets.githubusercontent.com/x": true,
		"https://RELEASE-ASSETS.githubusercontent.com/x": true,
		"http://objects.githubusercontent.com/x":         false,
		"https://api.github.com/x":                       false,
		"https://githubusercontent.com.evil.example/x":   false,
		"https://evil.example/x":                         false,
		"https://objects.githubusercontent.com:444/x":    false,
		"http://127.0.0.1:1234/x":                        false,
	} {
		u, _ := url.Parse(raw)
		if got := prod.CheckRedirectURL(u) == nil; got != want {
			t.Errorf("redirect %s allowed=%v, want %v", raw, got, want)
		}
	}
}

func TestCheckAPIURLAndEnvOverride(t *testing.T) {
	if err := (Policy{}).CheckAPIURL(ReleasesAPI); err != nil {
		t.Fatal(err)
	}
	if err := (Policy{}).CheckAPIURL("https://api.github.com/repos/evil/x/releases?per_page=30"); err == nil {
		t.Fatal("other repository accepted")
	}

	t.Setenv(EnvAPIOverride, "")
	if p := DefaultPolicy(); p.APIURL != ReleasesAPI || p.AllowLoopback {
		t.Fatalf("default policy = %#v", p)
	}
	t.Setenv(EnvAPIOverride, "http://127.0.0.1:9999/releases")
	if p := DefaultPolicy(); p.APIURL != "http://127.0.0.1:9999/releases" || !p.AllowLoopback {
		t.Fatalf("override policy = %#v", p)
	}
	t.Setenv(EnvAPIOverride, "https://evil.example/releases")
	if p := DefaultPolicy(); p.APIURL != ReleasesAPI || p.AllowLoopback {
		t.Fatalf("remote override must be ignored: %#v", p)
	}
}

func TestDownloadRejectsRedirectToForeignHost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://evil.example/payload.zip", http.StatusFound)
	}))
	defer srv.Close()
	var buf bytes.Buffer
	_, err := Policy{AllowLoopback: true}.download(t.Context(), srv.Client(), srv.URL+"/x.zip", 1<<20, -1, &buf, nil)
	if err == nil || !strings.Contains(err.Error(), "not allowed") || ErrorCode(err) != CodeNetwork {
		t.Fatalf("expected a rejected redirect, got %v", err)
	}
}

func TestDownloadEnforcesSize(t *testing.T) {
	body := bytes.Repeat([]byte("a"), 1000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chunked" {
			w.(http.Flusher).Flush() // no Content-Length
		}
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	p := Policy{AllowLoopback: true}
	var buf bytes.Buffer
	if _, err := p.download(t.Context(), srv.Client(), srv.URL+"/a", 1<<20, 999, &buf, nil); ErrorCode(err) != CodeBadSize {
		t.Fatalf("expected bad_size, got %v", err)
	}
	buf.Reset()
	if _, err := p.download(t.Context(), srv.Client(), srv.URL+"/chunked", 500, -1, &buf, nil); ErrorCode(err) != CodeBadSize {
		t.Fatalf("expected bad_size over the limit, got %v", err)
	}
	buf.Reset()
	if n, err := p.download(t.Context(), srv.Client(), srv.URL+"/chunked", 1<<20, 1000, &buf, nil); err != nil || n != 1000 {
		t.Fatalf("download = %d, %v", n, err)
	}
}
