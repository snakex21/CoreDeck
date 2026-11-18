package api

import (
	"encoding/json"
	"testing"
)

func TestUpdateBindingsJSONShapes(t *testing.T) {
	t.Setenv("COREDECK_UPDATE_API", "")
	a := &API{ExeDir: t.TempDir()}
	a.ConfigureUpdates(false)

	var settings map[string]any
	if err := json.Unmarshal([]byte(a.GetUpdateSettings()), &settings); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"channel", "autoCheck", "skippedVersion", "lastCheck"} {
		if _, ok := settings[k]; !ok {
			t.Errorf("get_update_settings lacks %s: %v", k, settings)
		}
	}
	if settings["channel"] != "stable" || settings["autoCheck"] != true {
		t.Errorf("defaults = %v", settings)
	}

	var ok map[string]any
	_ = json.Unmarshal([]byte(a.SetUpdateSettings(`{"channel":"test"}`)), &ok)
	if ok["success"] != true {
		t.Errorf("set_update_settings = %v", ok)
	}
	_ = json.Unmarshal([]byte(a.SetUpdateSettings(`{"channel":"bogus"}`)), &ok)
	if ok["error"] == nil {
		t.Errorf("invalid channel accepted: %v", ok)
	}

	var start map[string]any
	_ = json.Unmarshal([]byte(a.StartUpdate("not-a-version")), &start)
	if start["error"] == nil || start["errorCode"] == nil {
		t.Errorf("start_update = %v", start)
	}

	var prog map[string]any
	_ = json.Unmarshal([]byte(a.UpdateProgress("missing")), &prog)
	if prog["state"] != "error" {
		t.Errorf("update_progress = %v", prog)
	}

	var ready map[string]any
	_ = json.Unmarshal([]byte(a.UpdateReady()), &ready)
	if v, has := ready["rolledBack"]; !has || v != nil || ready["justUpdated"] != "" {
		t.Errorf("update_ready = %v", ready)
	}

	var rb map[string]any
	_ = json.Unmarshal([]byte(a.GetRollbackInfo()), &rb)
	if rb["available"] != false || rb["version"] != "" {
		t.Errorf("get_rollback_info = %v", rb)
	}
	_ = json.Unmarshal([]byte(a.RollbackUpdate()), &rb)
	if rb["error"] == nil {
		t.Errorf("rollback_update without backup = %v", rb)
	}

	var skip map[string]any
	_ = json.Unmarshal([]byte(a.SkipUpdateVersion("v2.9.0")), &skip)
	if skip["success"] != true {
		t.Errorf("skip_update_version = %v", skip)
	}
}
