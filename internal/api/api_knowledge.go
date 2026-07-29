package api

// Knowledge base read/write, bounded link checking and human link review.

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"launcher/internal/knowledge"
	"launcher/internal/linkcheck"
	"launcher/internal/linkreview"
)

// GetKnowledgeData returns the knowledge base JSON.
func (a *API) GetKnowledgeData() string {
	a.knowledgeMu.Lock()
	defer a.knowledgeMu.Unlock()

	if a.KnowledgeData == nil {
		if kd, err := knowledge.Load(a.WebDir); err == nil {
			a.KnowledgeData = kd
		} else if kd, err := knowledge.LoadFromFS(a.WebFS); err == nil {
			a.KnowledgeData = kd
		} else {
			slog.Error("load knowledge", "error", err)
			return `{"error": "Cannot load knowledge.json"}`
		}
	}
	data, _ := json.Marshal(a.KnowledgeData)
	return string(data)
}

// SaveKnowledgeData replaces the entire knowledge base.
func (a *API) SaveKnowledgeData(knowledgeJSON string) string {
	var kd knowledge.KnowledgeData
	if err := json.Unmarshal([]byte(knowledgeJSON), &kd); err != nil {
		return `{"error": "Invalid knowledge JSON"}`
	}
	a.knowledgeMu.Lock()
	defer a.knowledgeMu.Unlock()
	if err := knowledge.Save(a.WebDir, &kd); err != nil {
		return jsonError(err.Error())
	}
	a.KnowledgeData = &kd
	return `{"success": true}`
}

type knowledgeCheckItem struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Section  string `json:"section,omitempty"`
	Starred  bool   `json:"starred"`
}

type knowledgeCheckResult struct {
	Category   string `json:"category"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Section    string `json:"section,omitempty"`
	Starred    bool   `json:"starred"`
	State      string `json:"state"`
	StatusCode int    `json:"statusCode,omitempty"`
	FinalURL   string `json:"finalUrl,omitempty"`
	Detail     string `json:"detail"`
	CheckedAt  string `json:"checkedAt"`
}

type knowledgeCheckSession struct {
	mu      sync.Mutex
	total   int
	checked int
	done    bool
	results []knowledgeCheckResult
}

var knowledgeCheckSessions sync.Map
var knowledgeCheckCounter int64

// StartKnowledgeLinkCheck starts a user-requested, bounded technical check of
// links in one category. It only reports results; it never edits the catalogue.
func (a *API) StartKnowledgeLinkCheck(category string, limit int) string {
	category = strings.TrimSpace(category)
	if category == "" {
		return `{"error":"Wybierz kategorię do sprawdzenia"}`
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	a.knowledgeMu.Lock()
	if a.KnowledgeData == nil {
		if kd, err := knowledge.Load(a.WebDir); err == nil {
			a.KnowledgeData = kd
		} else if kd, err := knowledge.LoadFromFS(a.WebFS); err == nil {
			a.KnowledgeData = kd
		} else {
			a.knowledgeMu.Unlock()
			return jsonError(err.Error())
		}
	}
	items := make([]knowledgeCheckItem, 0, limit)
	for _, cat := range a.KnowledgeData.Categories {
		if cat.Name != category {
			continue
		}
		for _, link := range cat.Links {
			if len(items) >= limit {
				break
			}
			items = append(items, knowledgeCheckItem{
				Category: cat.Name, Name: link.Name, URL: link.URL, Section: link.Section, Starred: link.Starred,
			})
		}
		break
	}
	a.knowledgeMu.Unlock()

	if len(items) == 0 {
		return `{"error":"Nie znaleziono linków w tej kategorii"}`
	}

	sessionID := fmt.Sprintf("knowledge-check-%d", atomic.AddInt64(&knowledgeCheckCounter, 1))
	results := make([]knowledgeCheckResult, len(items))
	for index, item := range items {
		results[index] = knowledgeCheckResult{
			Category: item.Category, Name: item.Name, URL: item.URL, Section: item.Section, Starred: item.Starred,
			State: "pending", Detail: "Oczekuje w kolejce sprawdzania",
		}
	}
	session := &knowledgeCheckSession{total: len(items), results: results}
	knowledgeCheckSessions.Store(sessionID, session)

	go func() {
		jobs := make(chan int)
		var workers sync.WaitGroup
		workerCount := 4
		if len(items) < workerCount {
			workerCount = len(items)
		}
		for worker := 0; worker < workerCount; worker++ {
			workers.Add(1)
			go func() {
				defer workers.Done()
				for index := range jobs {
					item := items[index]
					check := linkcheck.Check(item.URL)
					result := knowledgeCheckResult{
						Category: item.Category, Name: item.Name, URL: item.URL, Section: item.Section, Starred: item.Starred,
						State: check.State, StatusCode: check.StatusCode, FinalURL: check.FinalURL, Detail: check.Detail, CheckedAt: check.CheckedAt,
					}
					session.mu.Lock()
					session.results[index] = result
					session.checked++
					session.mu.Unlock()
				}
			}()
		}
		for index := range items {
			jobs <- index
		}
		close(jobs)
		workers.Wait()
		session.mu.Lock()
		session.done = true
		session.mu.Unlock()
		time.AfterFunc(10*time.Minute, func() { knowledgeCheckSessions.Delete(sessionID) })
	}()

	data, _ := json.Marshal(map[string]interface{}{"sessionId": sessionID, "total": len(items)})
	return string(data)
}

// GetKnowledgeLinkCheck returns the current state of a running link check.
func (a *API) GetKnowledgeLinkCheck(sessionID string) string {
	value, ok := knowledgeCheckSessions.Load(sessionID)
	if !ok {
		return `{"error":"Nie znaleziono sesji sprawdzania","done":true}`
	}
	session := value.(*knowledgeCheckSession)
	session.mu.Lock()
	defer session.mu.Unlock()
	data, _ := json.Marshal(map[string]interface{}{
		"done": session.done, "total": session.total, "checked": session.checked, "results": session.results,
	})
	return string(data)
}

type knowledgeReviewRequest struct {
	Action    string `json:"action"`
	Category  string `json:"category"`
	URL       string `json:"url"`
	Confirmed bool   `json:"confirmed"`
}

// ReviewKnowledgeLink persists a human review decision. Remove is the only
// destructive action and must be explicitly confirmed for featured links.
func (a *API) ReviewKnowledgeLink(requestJSON string) string {
	var request knowledgeReviewRequest
	if err := json.Unmarshal([]byte(requestJSON), &request); err != nil {
		return `{"error":"Nieprawidłowa decyzja"}`
	}
	request.Action = strings.TrimSpace(request.Action)
	request.Category = strings.TrimSpace(request.Category)
	request.URL = strings.TrimSpace(request.URL)
	if request.Category == "" || request.URL == "" || (request.Action != "keep" && request.Action != "flag" && request.Action != "remove") {
		return `{"error":"Brakuje danych decyzji"}`
	}

	a.knowledgeMu.Lock()
	defer a.knowledgeMu.Unlock()
	if a.KnowledgeData == nil {
		if kd, err := knowledge.Load(a.WebDir); err == nil {
			a.KnowledgeData = kd
		} else if kd, err := knowledge.LoadFromFS(a.WebFS); err == nil {
			a.KnowledgeData = kd
		} else {
			return jsonError(err.Error())
		}
	}

	found := false
	featured := false
	if request.Action == "remove" {
		for catIndex := range a.KnowledgeData.Categories {
			cat := &a.KnowledgeData.Categories[catIndex]
			if cat.Name != request.Category {
				continue
			}
			for linkIndex, link := range cat.Links {
				if link.URL != request.URL {
					continue
				}
				found = true
				featured = link.Starred
				if featured && !request.Confirmed {
					return `{"error":"Wyróżniony link wymaga dodatkowego potwierdzenia","requiresConfirmation":true}`
				}
				cat.Links = append(cat.Links[:linkIndex], cat.Links[linkIndex+1:]...)
				break
			}
			break
		}
		if !found {
			return `{"error":"Link nie istnieje już w tej kategorii"}`
		}
	}

	reviews, err := knowledge.LoadReviewData(a.WebDir)
	if err != nil {
		return jsonError(err.Error())
	}
	reviews.Decisions[knowledge.ReviewKey(request.Category, request.URL)] = knowledge.ReviewDecision{
		Category: request.Category, URL: request.URL, Action: request.Action, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := knowledge.SaveReviewData(a.WebDir, reviews); err != nil {
		return jsonError(err.Error())
	}
	if request.Action == "remove" {
		if err := knowledge.Save(a.WebDir, a.KnowledgeData); err != nil {
			return jsonError(err.Error())
		}
	}
	data, _ := json.Marshal(map[string]interface{}{"success": true, "action": request.Action, "featured": featured})
	return string(data)
}

// GetKnowledgeLinkReviewDecision returns the most recently persisted human
// decision for one link. The main review queue polls this after opening the
// separate WebView2 inspection window, so its preview stays in sync.
func (a *API) GetKnowledgeLinkReviewDecision(category, rawURL string) string {
	category = strings.TrimSpace(category)
	rawURL = strings.TrimSpace(rawURL)
	if category == "" || rawURL == "" {
		return `{"error":"Brakuje kategorii lub adresu linku"}`
	}

	a.knowledgeMu.Lock()
	defer a.knowledgeMu.Unlock()
	reviews, err := knowledge.LoadReviewData(a.WebDir)
	if err != nil {
		return jsonError(err.Error())
	}
	decision, found := reviews.Decisions[knowledge.ReviewKey(category, rawURL)]
	data, _ := json.Marshal(map[string]interface{}{
		"found":    found,
		"decision": decision,
	})
	return string(data)
}

type knowledgeBulkKeepRequest struct {
	Category  string   `json:"category"`
	URLs      []string `json:"urls"`
	Confirmed bool     `json:"confirmed"`
}

// BulkKeepKnowledgeLinks records one explicit user approval for multiple
// existing links. The UI only offers it for technical-success results, while
// this method verifies that every approved URL still belongs to the category.
func (a *API) BulkKeepKnowledgeLinks(requestJSON string) string {
	var request knowledgeBulkKeepRequest
	if err := json.Unmarshal([]byte(requestJSON), &request); err != nil {
		return `{"error":"Nieprawidłowe zatwierdzenie zbiorcze"}`
	}
	request.Category = strings.TrimSpace(request.Category)
	if request.Category == "" || len(request.URLs) == 0 || !request.Confirmed {
		return `{"error":"Brakuje potwierdzenia lub linków"}`
	}

	a.knowledgeMu.Lock()
	defer a.knowledgeMu.Unlock()
	if a.KnowledgeData == nil {
		if kd, err := knowledge.Load(a.WebDir); err == nil {
			a.KnowledgeData = kd
		} else if kd, err := knowledge.LoadFromFS(a.WebFS); err == nil {
			a.KnowledgeData = kd
		} else {
			return jsonError(err.Error())
		}
	}

	existing := make(map[string]struct{})
	for _, category := range a.KnowledgeData.Categories {
		if category.Name != request.Category {
			continue
		}
		for _, link := range category.Links {
			existing[link.URL] = struct{}{}
		}
		break
	}

	reviews, err := knowledge.LoadReviewData(a.WebDir)
	if err != nil {
		return jsonError(err.Error())
	}
	updated := 0
	for _, rawURL := range request.URLs {
		rawURL = strings.TrimSpace(rawURL)
		if rawURL == "" {
			continue
		}
		if _, ok := existing[rawURL]; !ok {
			continue
		}
		reviews.Decisions[knowledge.ReviewKey(request.Category, rawURL)] = knowledge.ReviewDecision{
			Category: request.Category, URL: rawURL, Action: "keep", UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
		}
		updated++
	}
	if updated == 0 {
		return `{"error":"Żaden z linków nie istnieje już w tej kategorii"}`
	}
	if err := knowledge.SaveReviewData(a.WebDir, reviews); err != nil {
		return jsonError(err.Error())
	}
	data, _ := json.Marshal(map[string]interface{}{"success": true, "updated": updated})
	return string(data)
}

// OpenKnowledgeLinkReview opens a separate WebView2 review window for one
// link. Decisions made in its toolbar reuse the same protected review path as
// the in-page queue controls.
func (a *API) OpenKnowledgeLinkReview(category, rawURL string) string {
	category = strings.TrimSpace(category)
	rawURL = strings.TrimSpace(rawURL)
	if category == "" || rawURL == "" {
		return `{"error":"Brakuje kategorii lub adresu linku"}`
	}

	a.knowledgeMu.Lock()
	if a.KnowledgeData == nil {
		if kd, err := knowledge.Load(a.WebDir); err == nil {
			a.KnowledgeData = kd
		} else if kd, err := knowledge.LoadFromFS(a.WebFS); err == nil {
			a.KnowledgeData = kd
		} else {
			a.knowledgeMu.Unlock()
			return jsonError(err.Error())
		}
	}
	name := rawURL
	found := false
	for _, categoryData := range a.KnowledgeData.Categories {
		if categoryData.Name != category {
			continue
		}
		for _, link := range categoryData.Links {
			if link.URL == rawURL {
				name = link.Name
				found = true
				break
			}
		}
		break
	}
	a.knowledgeMu.Unlock()
	if !found {
		return `{"error":"Link nie istnieje już w tej kategorii"}`
	}

	openedAt := time.Now().UTC().Format(time.RFC3339Nano)
	err := linkreview.Open(linkreview.Config{
		Title: "Kontrola linku — " + name,
		URL:   rawURL,
		OnDecision: func(action string) string {
			request, _ := json.Marshal(knowledgeReviewRequest{
				Action: action, Category: category, URL: rawURL, Confirmed: action == "remove",
			})
			return a.ReviewKnowledgeLink(string(request))
		},
	})
	if err != nil {
		return jsonError(err.Error())
	}
	data, _ := json.Marshal(map[string]interface{}{"success": true, "openedAt": openedAt})
	return string(data)
}
