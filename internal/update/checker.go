package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	ChannelStable = "stable"
	ChannelTest   = "test"

	AssetZip  = "CoreDeck-Portable.zip"
	AssetSums = "SHA256SUMS.txt"
)

// Asset is a release asset as reported by the GitHub API.
type Asset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Release is the subset of a GitHub release the updater uses.
type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	HTMLURL     string  `json:"html_url"`
	Draft       bool    `json:"draft"`
	Prerelease  bool    `json:"prerelease"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`

	version Version
}

// Version returns the parsed tag (valid after FetchReleases/SelectRelease).
func (r *Release) Version() Version { return r.version }

// Asset finds an asset by exact name.
func (r *Release) Asset(name string) (Asset, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return Asset{}, false
}

// PageURL returns the release page, falling back to the tag URL when the API
// value is not a page of this repository.
func (r *Release) PageURL() string {
	if strings.HasPrefix(r.HTMLURL, ReleasePagePrefix) && !strings.ContainsAny(r.HTMLURL, " \"'<>") {
		return r.HTMLURL
	}
	return ReleasePagePrefix + "tag/v" + r.version.String()
}

// NormalizeChannel maps unknown values to the stable channel.
func NormalizeChannel(channel string) string {
	if strings.EqualFold(strings.TrimSpace(channel), ChannelTest) {
		return ChannelTest
	}
	return ChannelStable
}

// FetchReleases downloads and parses the release list. Releases with tags
// that are not semantic versions are dropped.
func FetchReleases(ctx context.Context, client *http.Client, policy Policy) ([]Release, error) {
	endpoint := policy.APIURL
	if endpoint == "" {
		endpoint = ReleasesAPI
	}
	if err := policy.CheckAPIURL(endpoint); err != nil {
		return nil, err
	}
	req, err := newRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := policy.apiClient(client).Do(req)
	if err != nil {
		return nil, wrapErr(CodeNetwork, fmt.Errorf("cannot check for updates: %w", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, newErr(CodeNetwork, "GitHub returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, MaxAPIResponse+1))
	if err != nil {
		return nil, wrapErr(CodeNetwork, fmt.Errorf("cannot read GitHub response: %w", err))
	}
	if int64(len(data)) > MaxAPIResponse {
		return nil, newErr(CodeNetwork, "GitHub response is too large")
	}
	return ParseReleases(data)
}

// ParseReleases decodes a GitHub release list and keeps valid version tags.
func ParseReleases(data []byte) ([]Release, error) {
	var raw []Release
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, newErr(CodeNetwork, "cannot decode GitHub response: %v", err)
	}
	out := make([]Release, 0, len(raw))
	for _, r := range raw {
		v, err := ParseVersion(r.TagName)
		if err != nil {
			continue
		}
		r.version = v
		out = append(out, r)
	}
	return out, nil
}

// eligible reports whether a release may be offered on the channel.
func eligible(r *Release, channel string) bool {
	if r.Draft {
		return false
	}
	if NormalizeChannel(channel) == ChannelStable && (r.Prerelease || r.version.IsPrerelease()) {
		return false
	}
	return true
}

// SelectRelease returns the highest release eligible for the channel, or nil.
func SelectRelease(releases []Release, channel string) *Release {
	var best *Release
	for i := range releases {
		r := &releases[i]
		if !eligible(r, channel) {
			continue
		}
		if best == nil || Compare(r.version, best.version) > 0 {
			best = r
		}
	}
	return best
}

// FindRelease returns the non-draft release with the given version that is
// eligible on the channel.
func FindRelease(releases []Release, channel, version string) *Release {
	want, err := ParseVersion(version)
	if err != nil {
		return nil
	}
	for i := range releases {
		r := &releases[i]
		if eligible(r, channel) && Compare(r.version, want) == 0 {
			return r
		}
	}
	return nil
}

// CheckResult is returned to the UI by check_for_updates.
type CheckResult struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	Prerelease      bool   `json:"prerelease"`
	Notes           string `json:"notes"`
	ReleaseURL      string `json:"releaseUrl"`
	PublishedAt     string `json:"publishedAt"`
	Channel         string `json:"channel"`
	Skipped         bool   `json:"skipped"`
	Error           string `json:"error,omitempty"`
	ErrorCode       string `json:"errorCode,omitempty"`
}

// Mode of a check: automatic (honours settings) or manual (ignores them).
const (
	ModeAuto   = "auto"
	ModeManual = "manual"
)

// AutoCheckInterval is the minimum time between automatic checks.
const AutoCheckInterval = 24 * time.Hour

// ShouldAutoCheck reports whether an automatic check may contact GitHub now.
func ShouldAutoCheck(s Settings, now time.Time) bool {
	if !s.AutoCheck {
		return false
	}
	if s.LastCheck == "" {
		return true
	}
	last, err := time.Parse(time.RFC3339, s.LastCheck)
	if err != nil {
		return true
	}
	if last.After(now) { // clock moved back: do not block checks forever
		return true
	}
	return now.Sub(last) >= AutoCheckInterval
}

// BuildResult fills a CheckResult from a release list.
func BuildResult(releases []Release, current string, s Settings, mode string) CheckResult {
	res := CheckResult{CurrentVersion: current, Channel: NormalizeChannel(s.Channel)}
	best := SelectRelease(releases, res.Channel)
	if best == nil {
		return res
	}
	res.LatestVersion = best.version.String()
	res.Prerelease = best.Prerelease || best.version.IsPrerelease()
	res.Notes = truncate(best.Body, MaxReleaseNotes)
	res.ReleaseURL = best.PageURL()
	if t, err := time.Parse(time.RFC3339, best.PublishedAt); err == nil {
		res.PublishedAt = t.UTC().Format(time.RFC3339)
	}
	if s.SkippedVersion != "" {
		if c, ok := CompareStrings(s.SkippedVersion, res.LatestVersion); ok && c == 0 {
			res.Skipped = true
		}
	}
	cur, err := ParseVersion(current)
	if err != nil {
		res.Error = fmt.Sprintf("invalid current version %q", current)
		res.ErrorCode = CodeInstall
		return res
	}
	newer := Compare(best.version, cur) > 0
	res.UpdateAvailable = newer && (mode == ModeManual || !res.Skipped)
	return res
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	s = s[:max]
	// do not cut a UTF-8 sequence in half
	for len(s) > 0 && s[len(s)-1]&0xC0 == 0x80 {
		s = s[:len(s)-1]
	}
	if len(s) > 0 && s[len(s)-1] >= 0xC0 {
		s = s[:len(s)-1]
	}
	return s + "\n…"
}
