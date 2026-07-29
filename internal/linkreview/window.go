// Package linkreview opens an isolated WebView2 window for a human to review
// one knowledge-base link. The page is never altered unless the user clicks a
// decision button in the native review toolbar.
package linkreview

import (
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/webview/webview_go"
)

type Config struct {
	Title      string
	URL        string
	OnDecision func(action string) string
}

// decisionSucceeded keeps the review window open when saving the user's
// decision failed. Before this guard, the window closed even after an API
// error, which made a failed confirmation indistinguishable from success.
func decisionSucceeded(result string) bool {
	var response struct {
		Success bool `json:"success"`
	}
	return json.Unmarshal([]byte(result), &response) == nil && response.Success
}

// Open creates a separate, native WebView2 review window without blocking the
// main launcher window. Only HTTP(S) locations are allowed.
func Open(cfg Config) error {
	parsed, err := url.Parse(cfg.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("review window requires an HTTP(S) URL")
	}
	if cfg.OnDecision == nil {
		return fmt.Errorf("review window requires a decision handler")
	}

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		window := webview.New(false)
		defer window.Destroy()
		window.SetTitle(cfg.Title)
		window.SetSize(1280, 820, webview.HintNone)
		window.Bind("__link_review_decide", func(action string) string {
			result := cfg.OnDecision(action)
			if decisionSucceeded(result) {
				// Closing only after a successful save avoids hiding an API error
				// behind a disappearing review window. Terminate must be dispatched
				// onto WebView's UI thread; calling it directly from the delayed
				// goroutine leaves the review window open on WebView2.
				go func() {
					time.Sleep(350 * time.Millisecond)
					window.Dispatch(func() {
						window.Terminate()
					})
				}()
			}
			return result
		})
		window.Init(toolbarScript)
		window.Navigate(cfg.URL)
		window.Run()
	}()
	return nil
}

// toolbarScript is injected into every navigation in the review window. It is
// deliberately self-contained and uses a shadow root so host-page styles do
// not hide the review controls. This is popup/notification hardening, not a
// replacement for a native EasyList request filter.
const toolbarScript = `(() => {
  try {
    window.open = () => null;
    if (window.Notification) window.Notification.requestPermission = () => Promise.resolve('denied');
  } catch (_) {}

  const mount = () => {
    if (document.getElementById('__launcher_link_review')) return;
    const host = document.createElement('div');
    host.id = '__launcher_link_review';
    host.style.cssText = 'position:fixed;z-index:2147483647;top:12px;right:12px;max-width:360px;';
    const root = host.attachShadow({mode:'open'});
    root.innerHTML = '<style>:host{all:initial} .box{font-family:Segoe UI,Arial,sans-serif;background:#12171d;color:#edf3f0;border:1px solid #38634d;border-radius:12px;padding:12px;box-shadow:0 12px 32px rgba(0,0,0,.45)}.title{font-size:13px;font-weight:700;margin-bottom:5px}.note{font-size:11px;color:#afbbb5;line-height:1.35;margin-bottom:10px}.actions{display:flex;gap:7px}.btn{border:1px solid #40514a;background:#202a25;color:#f2f7f4;border-radius:7px;padding:7px 9px;font-size:12px;font-weight:700;cursor:pointer}.btn:disabled{opacity:.55;cursor:wait}.keep{border-color:#29b86a;color:#7de6aa}.flag{border-color:#d1a62b;color:#ffd66b}.remove{border-color:#c34d4d;color:#ff9b9b}.status{display:none;margin-top:9px;font-size:11px;line-height:1.3}.status.ok{display:block;color:#7de6aa}.status.error{display:block;color:#ff9b9b}</style><div class="box"><div class="title">🩺 Kontrola linku</div><div class="note">Popupy i prośby o powiadomienia są blokowane. Sprawdź stronę, a potem podejmij decyzję.</div><div class="actions"><button class="btn keep" data-action="keep">✓ Zostaw</button><button class="btn flag" data-action="flag">⚠ Sprawdź później</button><button class="btn remove" data-action="remove">✕ Usuń</button></div><div class="status" role="status" aria-live="polite"></div></div>';
    root.querySelectorAll('[data-action]').forEach(button => button.addEventListener('click', async () => {
      const action = button.dataset.action;
      if (action === 'remove' && !window.confirm('Usunąć ten link z bazy wiedzy?')) return;
      const status = root.querySelector('.status');
      button.disabled = true;
      try {
        const raw = await __link_review_decide(action);
        const result = typeof raw === 'string' ? JSON.parse(raw) : raw;
        if (!result || !result.success) throw new Error(result && result.error ? result.error : 'Nie udało się zapisać decyzji.');
        status.className = 'status ok';
        status.textContent = '✓ Decyzja zapisana. Okno zostanie zamknięte.';
      } catch (error) {
        button.disabled = false;
        status.className = 'status error';
        status.textContent = 'Nie zapisano decyzji: ' + (error && error.message ? error.message : String(error));
      }
    }));
    (document.documentElement || document.body).appendChild(host);
  };
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', mount, {once:true}); else mount();
})();`
