// Package linkcheck performs conservative, user-triggered HTTP checks for
// knowledge-base links. It never changes links or decisions on its own.
package linkcheck

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"launcher/internal/uimsg"
)

const requestTimeout = 12 * time.Second

// Result contains the technical result of checking one HTTP(S) URL.
type Result struct {
	URL        string `json:"url"`
	State      string `json:"state"` // ok, warning, failed
	StatusCode int    `json:"statusCode,omitempty"`
	FinalURL   string `json:"finalUrl,omitempty"`
	Detail     string `json:"detail"`
	// DetailCode is the UI message key of Detail (translations.js), with
	// DetailParams as its placeholders; Detail is the Polish fallback.
	DetailCode   string            `json:"detailCode,omitempty"`
	DetailParams map[string]string `json:"detailParams,omitempty"`
	CheckedAt    string            `json:"checkedAt"`
}

func (r *Result) setDetail(m uimsg.Msg) {
	r.Detail, r.DetailCode, r.DetailParams = m.Text, m.Code, m.Params
}

// Check verifies one URL without downloading its body. HEAD is used first;
// servers that reject it are retried with a minimal GET request.
func Check(rawURL string) Result {
	result := Result{URL: rawURL, CheckedAt: time.Now().UTC().Format(time.RFC3339)}
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		result.State = "failed"
		result.setDetail(uimsg.New("msg_link_invalid_url", "Nieprawidłowy adres HTTP/HTTPS"))
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	client := &http.Client{
		Timeout: requestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	response, err := request(ctx, client, http.MethodHead, rawURL)
	if err == nil && (response.StatusCode == http.StatusMethodNotAllowed || response.StatusCode == http.StatusForbidden) {
		response.Body.Close()
		response, err = request(ctx, client, http.MethodGet, rawURL)
	}
	if err != nil {
		result.State = "failed"
		short := shortError(err)
		result.setDetail(uimsg.New("msg_link_request_failed", short, "error", short))
		return result
	}
	defer response.Body.Close()

	result.StatusCode = response.StatusCode
	result.FinalURL = response.Request.URL.String()
	switch {
	case response.StatusCode >= 200 && response.StatusCode < 400:
		result.State = "ok"
		if result.FinalURL != rawURL {
			result.setDetail(uimsg.New("msg_link_ok_redirect", "Działa po przekierowaniu"))
		} else {
			result.setDetail(uimsg.New("msg_link_ok", "Strona odpowiada poprawnie"))
		}
	case response.StatusCode >= 400 && response.StatusCode < 500:
		result.State = "warning"
		status := fmt.Sprintf("%d", response.StatusCode)
		result.setDetail(uimsg.New("msg_link_http_warning", "Serwer zwrócił HTTP "+status+" — warto otworzyć ręcznie", "status", status))
	default:
		result.State = "failed"
		status := fmt.Sprintf("%d", response.StatusCode)
		result.setDetail(uimsg.New("msg_link_http_failed", "Serwer zwrócił HTTP "+status, "status", status))
	}
	return result
}

func request(ctx context.Context, client *http.Client, method, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Launcher-Link-Validator/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	return client.Do(req)
}

func shortError(err error) string {
	message := strings.TrimSpace(err.Error())
	if len(message) > 180 {
		return message[:177] + "..."
	}
	return message
}
