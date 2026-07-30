package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	launcherupdate "launcher/internal/update"
	"launcher/internal/version"
)

func (a *API) GetAppVersion() string {
	return version.Current
}

func (a *API) CheckForUpdates() string {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	result := launcherupdate.Check(ctx, &http.Client{Timeout: 6 * time.Second}, launcherupdate.LatestReleaseAPI, version.Current)
	data, _ := json.Marshal(result)
	return string(data)
}
