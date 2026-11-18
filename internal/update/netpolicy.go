package update

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// RepoPath is the only repository the updater talks to.
	RepoPath = "snakex21/CoreDeck"
	// ReleasesAPI lists the newest releases (GitHub REST API).
	ReleasesAPI = "https://api.github.com/repos/" + RepoPath + "/releases?per_page=30"
	// DownloadPrefix is the only accepted prefix of a release asset URL.
	DownloadPrefix = "https://github.com/" + RepoPath + "/releases/download/"
	// ReleasePagePrefix is the accepted prefix of a release page (html_url).
	ReleasePagePrefix = "https://github.com/" + RepoPath + "/releases/"

	// EnvAPIOverride replaces ReleasesAPI for local testing against a fake
	// release server. Only http(s)://127.0.0.1[:port]/... is accepted; when it
	// is set, API requests and downloads may also use http://127.0.0.1.
	EnvAPIOverride = "COREDECK_UPDATE_API"

	// UserAgent is deliberately generic (no version or identifiers).
	UserAgent = "CoreDeck-updater"

	MaxDownloadSize      int64 = 200 << 20
	MaxSumsSize          int64 = 64 << 10
	MaxAPIResponse       int64 = 8 << 20
	MaxReleaseNotes            = 64 << 10
	maxRedirects               = 5
	loopbackHost               = "127.0.0.1"
	apiRequestTimeout          = 15 * time.Second
	downloadTotalTimeout       = 30 * time.Minute
)

// AllowedDownloadHosts are the hosts a download may be redirected to.
var AllowedDownloadHosts = []string{
	"github.com",
	"objects.githubusercontent.com",
	"release-assets.githubusercontent.com",
}

// Error codes reported to the UI.
const (
	CodeBadChecksum = "bad_checksum"
	CodeBadSize     = "bad_size"
	CodeNoAsset     = "no_asset"
	CodeNetwork     = "network"
	CodeInstall     = "install"
)

// Error is an updater error carrying a stable code for the UI.
type Error struct {
	Code string
	Err  error
}

func (e *Error) Error() string { return e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }

func newErr(code string, format string, args ...any) error {
	return &Error{Code: code, Err: fmt.Errorf(format, args...)}
}

func wrapErr(code string, err error) error {
	if err == nil {
		return nil
	}
	var ue *Error
	if errors.As(err, &ue) {
		return err
	}
	return &Error{Code: code, Err: err}
}

// ErrorCode returns the code of an updater error ("install" for anything else).
func ErrorCode(err error) string {
	var ue *Error
	if errors.As(err, &ue) && ue.Code != "" {
		return ue.Code
	}
	return CodeInstall
}

// Policy decides which URLs the updater may contact.
type Policy struct {
	APIURL        string
	AllowLoopback bool // only for a local fake release server (EnvAPIOverride)
}

// DefaultPolicy is the production policy, optionally overridden by
// EnvAPIOverride (loopback only).
func DefaultPolicy() Policy {
	p := Policy{APIURL: ReleasesAPI}
	if raw := strings.TrimSpace(os.Getenv(EnvAPIOverride)); raw != "" {
		if u, err := url.Parse(raw); err == nil && isLoopbackURL(u) {
			p.APIURL = raw
			p.AllowLoopback = true
		}
	}
	return p
}

func isLoopbackURL(u *url.URL) bool {
	if u == nil || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	return u.Hostname() == loopbackHost
}

// isAllowedHTTPS accepts https://<allowed host>[:443]/... without credentials.
func isAllowedHTTPS(u *url.URL, hosts []string) bool {
	if u == nil || u.Scheme != "https" || u.User != nil || u.Opaque != "" {
		return false
	}
	if port := u.Port(); port != "" && port != "443" {
		return false
	}
	host := u.Hostname()
	for _, h := range hosts {
		if strings.EqualFold(host, h) {
			return true
		}
	}
	return false
}

// CheckAPIURL validates the releases listing URL.
func (p Policy) CheckAPIURL(raw string) error {
	if p.AllowLoopback {
		if u, err := url.Parse(raw); err == nil && isLoopbackURL(u) {
			return nil
		}
	}
	if raw != ReleasesAPI {
		return newErr(CodeNetwork, "update API URL not allowed: %s", raw)
	}
	return nil
}

// CheckAssetURL validates a browser_download_url from the API.
func (p Policy) CheckAssetURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return newErr(CodeNoAsset, "invalid asset URL: %v", err)
	}
	if p.AllowLoopback && isLoopbackURL(u) {
		return nil
	}
	if !strings.HasPrefix(raw, DownloadPrefix) || !isAllowedHTTPS(u, []string{"github.com"}) ||
		strings.Contains(u.Path, "..") || u.RawQuery != "" || u.Fragment != "" {
		return newErr(CodeNoAsset, "asset URL not allowed: %s", raw)
	}
	return nil
}

// CheckRedirectURL validates every hop of a download redirect chain.
func (p Policy) CheckRedirectURL(u *url.URL) error {
	if p.AllowLoopback && isLoopbackURL(u) {
		return nil
	}
	if !isAllowedHTTPS(u, AllowedDownloadHosts) {
		return newErr(CodeNetwork, "redirect to a host that is not allowed: %s", u.Redacted())
	}
	return nil
}

// apiClient never follows redirects.
func (p Policy) apiClient(base *http.Client) *http.Client {
	c := cloneClient(base, apiRequestTimeout)
	c.CheckRedirect = func(*http.Request, []*http.Request) error {
		return newErr(CodeNetwork, "unexpected redirect from the update API")
	}
	return c
}

// downloadClient follows at most maxRedirects redirects to allowed hosts.
func (p Policy) downloadClient(base *http.Client) *http.Client {
	c := cloneClient(base, downloadTotalTimeout)
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) > maxRedirects {
			return newErr(CodeNetwork, "too many redirects")
		}
		return p.CheckRedirectURL(req.URL)
	}
	return c
}

func cloneClient(base *http.Client, timeout time.Duration) *http.Client {
	c := &http.Client{Timeout: timeout}
	if base != nil {
		c.Transport = base.Transport
		if base.Timeout > 0 {
			c.Timeout = base.Timeout
		}
	}
	c.Jar = nil // no cookies
	return c
}

func newRequest(ctx context.Context, raw string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, newErr(CodeNetwork, "invalid URL: %v", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	return req, nil
}

// download streams url into dst. maxBytes caps the body; expected (>=0) must
// match the final size exactly. progress may be nil.
func (p Policy) download(ctx context.Context, base *http.Client, raw string, maxBytes, expected int64, dst io.Writer, progress func(done, total int64)) (int64, error) {
	if err := p.CheckAssetURL(raw); err != nil {
		return 0, err
	}
	req, err := newRequest(ctx, raw)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/octet-stream")
	resp, err := p.downloadClient(base).Do(req)
	if err != nil {
		return 0, wrapErr(CodeNetwork, fmt.Errorf("download failed: %w", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, newErr(CodeNetwork, "download failed: HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > maxBytes {
		return 0, newErr(CodeBadSize, "file is too large (%d bytes)", resp.ContentLength)
	}
	if expected >= 0 && resp.ContentLength >= 0 && resp.ContentLength != expected {
		return 0, newErr(CodeBadSize, "unexpected size: %d bytes, expected %d", resp.ContentLength, expected)
	}
	total := expected
	if total < 0 {
		total = resp.ContentLength
	}
	limit := maxBytes
	if expected >= 0 && expected < limit {
		limit = expected
	}
	body := &idleReader{r: io.LimitReader(resp.Body, limit+1), ctx: ctx}
	var done int64
	buf := make([]byte, 64<<10)
	for {
		n, rerr := body.Read(buf)
		if n > 0 {
			if done+int64(n) > limit {
				return done, newErr(CodeBadSize, "download exceeds the expected size")
			}
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return done, wrapErr(CodeInstall, werr)
			}
			done += int64(n)
			if progress != nil {
				progress(done, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return done, wrapErr(CodeNetwork, fmt.Errorf("download interrupted: %w", rerr))
		}
	}
	if expected >= 0 && done != expected {
		return done, newErr(CodeBadSize, "downloaded %d bytes, expected %d", done, expected)
	}
	return done, nil
}

// idleReader only forwards context cancellation; the overall timeout is the
// client's. Kept separate so a stalled body surfaces as a network error.
type idleReader struct {
	r   io.Reader
	ctx context.Context
}

func (r *idleReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}
