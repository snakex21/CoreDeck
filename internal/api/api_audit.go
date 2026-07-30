package api

import "log/slog"

func (a *API) recordAudit(action, target, status, sessionID string, details map[string]any) {
	if a == nil || a.AuditLog == nil {
		return
	}
	if err := a.AuditLog.Record(action, target, status, sessionID, details); err != nil {
		slog.Warn("audit log write failed", "action", action, "error", err)
	}
}
