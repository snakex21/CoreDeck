package api

import (
	"encoding/json"
	"time"

	"launcher/internal/uimsg"
)

func (a *API) EnsureLLMStarted() string {
	if a.LLMClient == nil {
		return jsonMsg(uimsg.New("msg_ai_unavailable", "Lokalny moduł AI nie jest dostępny"))
	}
	if err := a.LLMClient.EnsureReady(a.ExeDir, 25*time.Second); err != nil {
		return jsonErr(err)
	}
	data, _ := json.Marshal(map[string]any{
		"ready":   true,
		"port":    a.LLMClient.Port,
		"baseUrl": a.LLMClient.BaseURL,
	})
	return string(data)
}
