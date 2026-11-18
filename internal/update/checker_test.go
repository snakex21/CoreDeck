package update

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const releasesFixture = `[
  {"tag_name":"v2.6.0","draft":true,"prerelease":false,"body":"draft"},
  {"tag_name":"v2.5.0-rc.2","draft":false,"prerelease":true,"body":"rc2","published_at":"2026-11-10T10:00:00Z"},
  {"tag_name":"v2.5.0-rc.10","draft":false,"prerelease":true,"body":"rc10"},
  {"tag_name":"v2.5.0-beta.3","draft":false,"prerelease":true},
  {"tag_name":"nightly","draft":false,"prerelease":true},
  {"tag_name":"v2.4.1","draft":false,"prerelease":false,"body":"stable notes",
   "html_url":"https://github.com/snakex21/CoreDeck/releases/tag/v2.4.1","published_at":"2026-11-01T08:00:00+02:00"},
  {"tag_name":"v2.4.0","draft":false,"prerelease":false},
  {"tag_name":"v2.7.0-rc.1","draft":true,"prerelease":true}
]`

func mustParse(t *testing.T, data string) []Release {
	t.Helper()
	rels, err := ParseReleases([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	return rels
}

func TestSelectReleaseByChannel(t *testing.T) {
	rels := mustParse(t, releasesFixture)
	if len(rels) != 7 {
		t.Fatalf("invalid tags must be dropped, got %d releases", len(rels))
	}
	if r := SelectRelease(rels, ChannelStable); r == nil || r.Version().String() != "2.4.1" {
		t.Fatalf("stable = %v", r)
	}
	if r := SelectRelease(rels, ChannelTest); r == nil || r.Version().String() != "2.5.0-rc.10" {
		t.Fatalf("test = %v (rc.10 > rc.2, drafts ignored)", r)
	}
	if r := SelectRelease(rels, "unknown"); r == nil || r.Version().String() != "2.4.1" {
		t.Fatalf("unknown channel must behave as stable, got %v", r)
	}

	// A stable release beats its own prereleases on the test channel.
	withFinal := mustParse(t, `[{"tag_name":"v2.5.0-rc.10","prerelease":true},{"tag_name":"v2.5.0"}]`)
	if r := SelectRelease(withFinal, ChannelTest); r.Version().String() != "2.5.0" {
		t.Fatalf("test channel picked %s", r.Version())
	}
	// A tag with a prerelease suffix is never offered on stable even when the
	// GitHub flag was forgotten.
	mislabeled := mustParse(t, `[{"tag_name":"v9.0.0-rc.1","prerelease":false},{"tag_name":"v2.4.1"}]`)
	if r := SelectRelease(mislabeled, ChannelStable); r.Version().String() != "2.4.1" {
		t.Fatalf("stable picked %s", r.Version())
	}
	if FindRelease(rels, ChannelStable, "2.6.0") != nil {
		t.Fatal("draft must not be found")
	}
	if FindRelease(rels, ChannelStable, "2.5.0-rc.2") != nil {
		t.Fatal("prerelease must not be found on stable")
	}
	if FindRelease(rels, ChannelTest, "v2.5.0-rc.2") == nil {
		t.Fatal("prerelease must be found on test")
	}
}

func TestBuildResult(t *testing.T) {
	rels := mustParse(t, releasesFixture)
	s := DefaultSettings()

	res := BuildResult(rels, "2.4.0", s, ModeAuto)
	if !res.UpdateAvailable || res.LatestVersion != "2.4.1" || res.Prerelease || res.Notes != "stable notes" {
		t.Fatalf("unexpected %#v", res)
	}
	if res.ReleaseURL != "https://github.com/snakex21/CoreDeck/releases/tag/v2.4.1" || res.PublishedAt != "2026-11-01T06:00:00Z" {
		t.Fatalf("unexpected url/date %#v", res)
	}

	s.SkippedVersion = "2.4.1"
	if res := BuildResult(rels, "2.4.0", s, ModeAuto); res.UpdateAvailable || !res.Skipped {
		t.Fatalf("auto check must honour the skipped version: %#v", res)
	}
	if res := BuildResult(rels, "2.4.0", s, ModeManual); !res.UpdateAvailable || !res.Skipped {
		t.Fatalf("manual check must ignore the skipped version: %#v", res)
	}

	if res := BuildResult(rels, "2.4.1", DefaultSettings(), ModeManual); res.UpdateAvailable {
		t.Fatalf("same version must not be offered: %#v", res)
	}
	if res := BuildResult(rels, "2.9.0", DefaultSettings(), ModeManual); res.UpdateAvailable {
		t.Fatalf("no downgrade: %#v", res)
	}
	test := Settings{Channel: ChannelTest, AutoCheck: true}
	if res := BuildResult(rels, "2.4.1", test, ModeAuto); !res.UpdateAvailable || !res.Prerelease || res.Channel != ChannelTest {
		t.Fatalf("test channel: %#v", res)
	}
	// html_url pointing elsewhere is replaced by the tag page.
	odd := mustParse(t, `[{"tag_name":"v3.0.0","html_url":"https://evil.example/x"}]`)
	if res := BuildResult(odd, "2.4.0", DefaultSettings(), ModeManual); res.ReleaseURL != "https://github.com/snakex21/CoreDeck/releases/tag/v3.0.0" {
		t.Fatalf("ReleaseURL = %q", res.ReleaseURL)
	}
}

func TestShouldAutoCheck(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	s := DefaultSettings()
	if !ShouldAutoCheck(s, now) {
		t.Fatal("first check must run")
	}
	s.LastCheck = now.Add(-23 * time.Hour).Format(time.RFC3339)
	if ShouldAutoCheck(s, now) {
		t.Fatal("must wait 24 h")
	}
	s.LastCheck = now.Add(-25 * time.Hour).Format(time.RFC3339)
	if !ShouldAutoCheck(s, now) {
		t.Fatal("24 h passed")
	}
	s.LastCheck = now.Add(48 * time.Hour).Format(time.RFC3339)
	if !ShouldAutoCheck(s, now) {
		t.Fatal("a future timestamp must not block checks")
	}
	s.AutoCheck = false
	s.LastCheck = ""
	if ShouldAutoCheck(s, now) {
		t.Fatal("disabled")
	}
}

func TestFetchReleases(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != UserAgent {
			t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
		}
		_, _ = w.Write([]byte(releasesFixture))
	}))
	defer srv.Close()

	rels, err := FetchReleases(t.Context(), srv.Client(), Policy{APIURL: srv.URL, AllowLoopback: true})
	if err != nil || len(rels) != 7 {
		t.Fatalf("FetchReleases = %d, %v", len(rels), err)
	}
	// Without the loopback override only the GitHub API URL is accepted.
	if _, err := FetchReleases(t.Context(), srv.Client(), Policy{APIURL: srv.URL}); ErrorCode(err) != CodeNetwork {
		t.Fatalf("expected a rejected API URL, got %v", err)
	}

	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "rate limited", http.StatusForbidden)
	}))
	defer failing.Close()
	if _, err := FetchReleases(t.Context(), failing.Client(), Policy{APIURL: failing.URL, AllowLoopback: true}); ErrorCode(err) != CodeNetwork {
		t.Fatalf("expected network error, got %v", err)
	}
}
