package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	LatestReleaseAPI = "https://api.github.com/repos/snakex21/CoreDeck/releases/latest"
	releaseBaseURL   = "https://github.com/snakex21/CoreDeck/releases/tag/"
)

var stableVersionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

type Result struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion,omitempty"`
	UpdateAvailable bool   `json:"updateAvailable"`
	ReleaseURL      string `json:"releaseUrl,omitempty"`
	PublishedAt     string `json:"publishedAt,omitempty"`
	Error           string `json:"error,omitempty"`
}

type githubRelease struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
}

func Check(ctx context.Context, client *http.Client, endpoint, current string) Result {
	result := Result{CurrentVersion: current}
	if client == nil {
		client = &http.Client{Timeout: 6 * time.Second}
	}
	if strings.TrimSpace(endpoint) == "" {
		endpoint = LatestReleaseAPI
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "CoreDeck/"+current)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	response, err := client.Do(request)
	if err != nil {
		result.Error = fmt.Sprintf("cannot check for updates: %v", err)
		return result
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		result.Error = fmt.Sprintf("GitHub returned HTTP %d", response.StatusCode)
		return result
	}

	var release githubRelease
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(&release); err != nil {
		result.Error = fmt.Sprintf("cannot decode GitHub response: %v", err)
		return result
	}

	latest, err := parseStableVersion(release.TagName)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	currentParts, err := parseStableVersion(current)
	if err != nil {
		result.Error = fmt.Sprintf("invalid current version: %v", err)
		return result
	}

	result.LatestVersion = strings.TrimPrefix(release.TagName, "v")
	result.UpdateAvailable = compareVersion(latest, currentParts) > 0
	result.ReleaseURL = releaseBaseURL + "v" + result.LatestVersion
	if !release.PublishedAt.IsZero() {
		result.PublishedAt = release.PublishedAt.UTC().Format(time.RFC3339)
	}
	return result
}

func parseStableVersion(raw string) ([3]int, error) {
	var parts [3]int
	match := stableVersionPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if match == nil {
		return parts, fmt.Errorf("invalid stable release version %q", raw)
	}
	for i := range parts {
		value, err := strconv.Atoi(match[i+1])
		if err != nil {
			return parts, fmt.Errorf("invalid stable release version %q", raw)
		}
		parts[i] = value
	}
	return parts, nil
}

func compareVersion(left, right [3]int) int {
	for i := range left {
		switch {
		case left[i] > right[i]:
			return 1
		case left[i] < right[i]:
			return -1
		}
	}
	return 0
}
