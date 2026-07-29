package api

import (
	"encoding/json"
	"time"
)

func (a *API) EnsureLLMStarted() string {
	if a.LLMClient == nil {
		return `{"error":"Lokalny moduł AI nie jest dostępny"}`
	}
	if err := a.LLMClient.EnsureReady(a.ExeDir, 25*time.Second); err != nil {
		return jsonError(err.Error())
	}
	data, _ := json.Marshal(map[string]any{
		"ready":   true,
		"port":    a.LLMClient.Port,
		"baseUrl": a.LLMClient.BaseURL,
	})
	return string(data)
}
