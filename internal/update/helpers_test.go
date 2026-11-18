package update

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type zipEntry struct {
	name string
	data []byte
}

func makeZip(t *testing.T, entries ...zipEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: e.name, Method: zip.Deflate})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(e.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sha(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

var (
	newExe    = []byte("NEW CoreDeck.exe 2.5.0")
	newHelper = []byte("NEW disk-helper.exe")
	oldExe    = []byte("OLD CoreDeck.exe 2.4.0")
	oldHelper = []byte("OLD disk-helper.exe")
)

// goodZip mirrors the real package: backslash separators, release-manifest.
func goodZip(t *testing.T) []byte {
	return makeZip(t,
		zipEntry{"CoreDeck.exe", newExe},
		zipEntry{`data\tools\disk-helper.exe`, newHelper},
		zipEntry{"release-manifest.json", []byte(`{"product":"CoreDeck"}`)},
	)
}

func sumsFor(zipData []byte) string {
	return fmt.Sprintf("%s  CoreDeck-Portable.zip\n%s  CoreDeck.exe\n%s  disk-helper.exe\n",
		sha(zipData), sha(newExe), sha(newHelper))
}

// newAppDir creates an installed CoreDeck 2.4.0 with some user data.
func newAppDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "CoreDeck.exe"), oldExe)
	writeFile(t, filepath.Join(dir, "data", "tools", "disk-helper.exe"), oldHelper)
	writeFile(t, filepath.Join(dir, "data", "settings.json"), []byte(`{"user":true}`))
	writeFile(t, filepath.Join(dir, "data", "programs.json"), []byte(`[]`))
	return dir
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// snapshot lists every file outside the work directory with its content.
func snapshot(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		if d.IsDir() {
			if rel == WorkDirName {
				return filepath.SkipDir
			}
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func assertSnapshot(t *testing.T, want, got map[string]string) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("file set changed:\nwant %v\ngot  %v", keys(want), keys(got))
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("file %s changed: %q -> %q", k, v, got[k])
		}
	}
}

func keys(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// fakeRelease serves a releases API, the zip and SHA256SUMS.txt on 127.0.0.1.
type fakeRelease struct {
	srv       *httptest.Server
	zip       []byte
	sums      string
	assetSize int64 // size reported by the API (0 = len(zip))
	releases  func(base string) []map[string]any
	hits      map[string]int
}

func newFakeRelease(t *testing.T, zipData []byte, sums string) *fakeRelease {
	t.Helper()
	f := &fakeRelease{zip: zipData, sums: sums, hits: map[string]int{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/releases", func(w http.ResponseWriter, r *http.Request) {
		f.hits["api"]++
		size := f.assetSize
		if size == 0 {
			size = int64(len(f.zip))
		}
		base := f.srv.URL
		var rels []map[string]any
		if f.releases != nil {
			rels = f.releases(base)
		} else {
			rels = []map[string]any{{
				"tag_name": "v2.5.0", "body": "notes", "draft": false, "prerelease": false,
				"published_at": "2026-11-01T10:00:00Z",
				"html_url":     "https://github.com/snakex21/CoreDeck/releases/tag/v2.5.0",
				"assets": []map[string]any{
					{"name": AssetZip, "size": size, "browser_download_url": base + "/dl/v2.5.0/" + AssetZip},
					{"name": AssetSums, "size": len(f.sums), "browser_download_url": base + "/dl/v2.5.0/" + AssetSums},
				},
			}}
		}
		_ = json.NewEncoder(w).Encode(rels)
	})
	mux.HandleFunc("/dl/", func(w http.ResponseWriter, r *http.Request) {
		f.hits[r.URL.Path]++
		switch {
		case strings.HasSuffix(r.URL.Path, AssetZip):
			_, _ = w.Write(f.zip)
		case strings.HasSuffix(r.URL.Path, AssetSums):
			_, _ = w.Write([]byte(f.sums))
		default:
			http.NotFound(w, r)
		}
	})
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeRelease) policy() Policy {
	return Policy{APIURL: f.srv.URL + "/api/releases", AllowLoopback: true}
}

func (f *fakeRelease) release(t *testing.T) *Release {
	t.Helper()
	rels, err := FetchReleases(t.Context(), f.srv.Client(), f.policy())
	if err != nil {
		t.Fatal(err)
	}
	if len(rels) == 0 {
		t.Fatal("no releases")
	}
	return &rels[0]
}
