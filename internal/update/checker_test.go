package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckFindsNewerStableRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "CoreDeck/2.1.1" {
			t.Errorf("User-Agent = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name":"v2.2.0",
			"published_at":"2026-07-30T10:00:00Z"
		}`))
	}))
	defer server.Close()

	got := Check(context.Background(), server.Client(), server.URL, "2.1.1")
	if got.Error != "" {
		t.Fatalf("Check() error = %q", got.Error)
	}
	if !got.UpdateAvailable || got.LatestVersion != "2.2.0" {
		t.Fatalf("unexpected result: %#v", got)
	}
	if got.ReleaseURL != "https://github.com/snakex21/CoreDeck/releases/tag/v2.2.0" {
		t.Fatalf("ReleaseURL = %q", got.ReleaseURL)
	}
}

func TestCheckDoesNotDowngradeOrReinstall(t *testing.T) {
	for _, tag := range []string{"v2.1.1", "v2.0.9"} {
		t.Run(tag, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"tag_name":"` + tag + `"}`))
			}))
			defer server.Close()

			got := Check(context.Background(), server.Client(), server.URL, "2.1.1")
			if got.Error != "" || got.UpdateAvailable {
				t.Fatalf("unexpected result: %#v", got)
			}
		})
	}
}

func TestCheckReportsGitHubAndVersionErrors(t *testing.T) {
	t.Run("http error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "rate limited", http.StatusForbidden)
		}))
		defer server.Close()

		got := Check(context.Background(), server.Client(), server.URL, "2.1.1")
		if got.Error == "" {
			t.Fatalf("expected an error, got %#v", got)
		}
	})

	t.Run("non-stable tag", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"tag_name":"v2.2.0-beta.1"}`))
		}))
		defer server.Close()

		got := Check(context.Background(), server.Client(), server.URL, "2.1.1")
		if got.Error == "" {
			t.Fatalf("expected an error, got %#v", got)
		}
	})
}
