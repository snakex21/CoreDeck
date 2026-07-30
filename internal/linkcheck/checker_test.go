package linkcheck

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestCheckRejectsNonHTTPURLs(t *testing.T) {
	for _, rawURL := range []string{"", "example.com", "file:///C:/Windows", "javascript:alert(1)"} {
		t.Run(rawURL, func(t *testing.T) {
			got := Check(rawURL)
			if got.State != "failed" || got.StatusCode != 0 ||
				!strings.Contains(got.Detail, "HTTP/HTTPS") {
				t.Fatalf("Check(%q) = %#v", rawURL, got)
			}
		})
	}
}

func TestCheckReportsSuccessfulResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("method = %s, want HEAD", r.Method)
		}
		if got := r.Header.Get("User-Agent"); got != "Launcher-Link-Validator/1.0" {
			t.Errorf("User-Agent = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	got := Check(server.URL)
	if got.State != "ok" || got.StatusCode != http.StatusNoContent ||
		got.FinalURL != server.URL || got.Detail == "" || got.CheckedAt == "" {
		t.Fatalf("Check() = %#v", got)
	}
}

func TestCheckFallsBackToMinimalGetWhenHeadIsRejected(t *testing.T) {
	headCalls := 0
	getCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			headCalls++
			w.WriteHeader(http.StatusMethodNotAllowed)
		case http.MethodGet:
			getCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("body is intentionally ignored"))
		default:
			t.Errorf("unexpected method: %s", r.Method)
		}
	}))
	defer server.Close()

	got := Check(server.URL)
	if got.State != "ok" || got.StatusCode != http.StatusOK {
		t.Fatalf("Check() = %#v", got)
	}
	if headCalls != 1 || getCalls != 1 {
		t.Fatalf("HEAD calls = %d, GET calls = %d; want 1 each", headCalls, getCalls)
	}
}

func TestCheckReportsRedirectAndFinalURL(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusFound)
	})
	mux.HandleFunc("/final", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	got := Check(server.URL + "/start")
	if got.State != "ok" || got.FinalURL != server.URL+"/final" ||
		!strings.Contains(got.Detail, "przekierowaniu") {
		t.Fatalf("Check() = %#v", got)
	}
}

func TestCheckClassifiesHTTPFailures(t *testing.T) {
	tests := []struct {
		status int
		state  string
	}{
		{status: http.StatusNotFound, state: "warning"},
		{status: http.StatusTooManyRequests, state: "warning"},
		{status: http.StatusInternalServerError, state: "failed"},
	}
	for _, test := range tests {
		t.Run(http.StatusText(test.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
			}))
			defer server.Close()

			got := Check(server.URL)
			if got.State != test.state || got.StatusCode != test.status ||
				!strings.Contains(got.Detail, strconv.Itoa(test.status)) {
				t.Fatalf("Check() = %#v, want state %q and status %d", got, test.state, test.status)
			}
		})
	}
}

func TestShortErrorBoundsDiagnosticLength(t *testing.T) {
	message := strings.Repeat("x", 300)
	got := shortError(errors.New(message))
	if len(got) != 180 || !strings.HasSuffix(got, "...") {
		t.Fatalf("shortError() length = %d, value = %q", len(got), got)
	}
	if got := shortError(errors.New(" short ")); got != "short" {
		t.Fatalf("shortError() = %q, want trimmed message", got)
	}
}
