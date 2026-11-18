package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Settings are the user's update preferences (data\state\update.json).
type Settings struct {
	Channel        string `json:"channel"`
	AutoCheck      bool   `json:"autoCheck"`
	SkippedVersion string `json:"skippedVersion"`
	LastCheck      string `json:"lastCheck"` // RFC 3339, UTC
}

// DefaultSettings: stable channel, automatic check enabled.
func DefaultSettings() Settings {
	return Settings{Channel: ChannelStable, AutoCheck: true}
}

// SettingsPath is <exeDir>\data\state\update.json.
func SettingsPath(exeDir string) string {
	return filepath.Join(exeDir, "data", "state", "update.json")
}

// LoadSettings returns defaults for a missing or unreadable file.
func LoadSettings(path string) Settings {
	s := DefaultSettings()
	data, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	var raw struct {
		Channel        *string `json:"channel"`
		AutoCheck      *bool   `json:"autoCheck"`
		SkippedVersion *string `json:"skippedVersion"`
		LastCheck      *string `json:"lastCheck"`
	}
	if json.Unmarshal(data, &raw) != nil {
		return s
	}
	if raw.Channel != nil {
		s.Channel = NormalizeChannel(*raw.Channel)
	}
	if raw.AutoCheck != nil {
		s.AutoCheck = *raw.AutoCheck
	}
	if raw.SkippedVersion != nil {
		if v, err := ParseVersion(*raw.SkippedVersion); err == nil {
			s.SkippedVersion = v.String()
		}
	}
	if raw.LastCheck != nil {
		s.LastCheck = *raw.LastCheck
	}
	return s
}

// SaveSettings writes the file atomically.
func SaveSettings(path string, s Settings) error {
	s.Channel = NormalizeChannel(s.Channel)
	return writeJSONAtomic(path, s)
}

// SettingsPatch is what the UI may change (lastCheck is managed internally).
type SettingsPatch struct {
	Channel        *string `json:"channel"`
	AutoCheck      *bool   `json:"autoCheck"`
	SkippedVersion *string `json:"skippedVersion"`
}

// ApplyPatch validates and merges a JSON patch from the UI.
func ApplyPatch(s Settings, raw string) (Settings, error) {
	// Unknown fields (e.g. lastCheck echoed back from get_update_settings)
	// are ignored.
	var p SettingsPatch
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return s, fmt.Errorf("invalid settings: %v", err)
	}
	if p.Channel != nil {
		switch c := strings.ToLower(strings.TrimSpace(*p.Channel)); c {
		case ChannelStable, ChannelTest:
			s.Channel = c
		default:
			return s, fmt.Errorf("unknown channel %q", *p.Channel)
		}
	}
	if p.AutoCheck != nil {
		s.AutoCheck = *p.AutoCheck
	}
	if p.SkippedVersion != nil {
		v, err := normalizeOptionalVersion(*p.SkippedVersion)
		if err != nil {
			return s, err
		}
		s.SkippedVersion = v
	}
	return s, nil
}

func normalizeOptionalVersion(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	v, err := ParseVersion(raw)
	if err != nil {
		return "", err
	}
	return v.String(), nil
}

// writeJSONAtomic writes through a temporary file in the same directory.
func writeJSONAtomic(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(path, data)
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	_, werr := tmp.Write(data)
	if werr == nil {
		werr = tmp.Sync()
	}
	if cerr := tmp.Close(); werr == nil {
		werr = cerr
	}
	if werr == nil {
		werr = moveFile(tmpPath, path)
	}
	if werr != nil {
		_ = os.Remove(tmpPath)
		return werr
	}
	return nil
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("empty file")
	}
	return json.Unmarshal(data, v)
}
