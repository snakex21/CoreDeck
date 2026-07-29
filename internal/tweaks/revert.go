package tweaks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var stateDir = "."

// SetStateDir configures where the tweak backup file is stored.
func SetStateDir(dir string) {
	if dir != "" {
		stateDir = dir
	}
}

func backupPath() string {
	return filepath.Join(stateDir, ".launcher_tweaks_backup.json")
}

// ValueBackup records the prior state of a single registry value before a tweak
// overwrote it, so revert can restore it (or delete it if it didn't exist).
type ValueBackup struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Existed bool   `json:"existed"`
	Type    string `json:"type,omitempty"`
	Data    string `json:"data,omitempty"`
}

// ServiceBackup records the prior start type of a service.
type ServiceBackup struct {
	Service   string `json:"service"`
	StartType string `json:"startType"`
}

// TweakBackup is the snapshot stored per applied tweak.
type TweakBackup struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Values    []ValueBackup   `json:"values,omitempty"`
	Services  []ServiceBackup `json:"services,omitempty"`
	AppliedAt string          `json:"appliedAt"`
}

func loadBackups() map[string]TweakBackup {
	m := map[string]TweakBackup{}
	data, err := os.ReadFile(backupPath())
	if err != nil {
		return m
	}
	if err := json.Unmarshal(data, &m); err != nil || m == nil {
		return map[string]TweakBackup{}
	}
	return m
}

func saveBackups(m map[string]TweakBackup) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(backupPath(), data, 0644)
}

// IsRevertible reports whether a tweak type supports automatic revert.
func IsRevertible(t Tweak) bool {
	return t.Type == "registry" || t.Type == "service"
}

// captureBackup snapshots the current system state for a tweak BEFORE it is
// applied, so it can be reverted later. Only registry and service tweaks are
// captured. An existing snapshot for the same id is never overwritten (so the
// original pre-tweak state is preserved across repeated applies).
func captureBackup(t Tweak) {
	switch t.Type {
	case "registry":
		b := TweakBackup{ID: t.ID, Type: "registry", AppliedAt: time.Now().Format(time.RFC3339)}
		for _, c := range splitCommands(t.Command) {
			c = strings.TrimSpace(c)
			if c == "" {
				continue
			}
			// A registry-typed tweak may embed a service change (sc config).
			// Capture the current start type so revert can restore it too.
			if sm := scConfigRe.FindStringSubmatch(c); len(sm) == 3 {
				svc := sm[1]
				if st, ok := readServiceStartType(svc); ok {
					b.Services = append(b.Services, ServiceBackup{Service: svc, StartType: st})
				}
				continue
			}
			// reg delete: capture the current value (if any) so revert can re-add
			// it. If it doesn't currently exist, store nothing — nothing to restore.
			if dm := regDeleteRe.FindStringSubmatch(c); len(dm) == 3 {
				path := dm[1]
				rest := dm[2]
				name := ""
				if strings.Contains(strings.ToLower(rest), "/ve") {
					name = "(Default)"
				} else if vm := regValueRe.FindStringSubmatch(rest); len(vm) == 2 {
					name = strings.Trim(vm[1], `"`)
				}
				if name == "" {
					continue
				}
				if curType, curData, ok := readRegValue(path, name); ok {
					b.Values = append(b.Values, ValueBackup{Path: path, Name: name, Existed: true, Type: curType, Data: curData})
				}
				continue
			}
			m := regAddRe.FindStringSubmatch(c)
			if len(m) != 3 {
				continue
			}
			path := m[1]
			rest := m[2]
			name := ""
			if strings.Contains(strings.ToLower(rest), "/ve") {
				name = "(Default)"
			} else if vm := regValueRe.FindStringSubmatch(rest); len(vm) == 2 {
				name = strings.Trim(vm[1], `"`)
			}
			if name == "" {
				continue
			}
			vb := ValueBackup{Path: path, Name: name}
			if curType, curData, ok := readRegValue(path, name); ok {
				vb.Existed = true
				vb.Type = curType
				vb.Data = curData
			}
			b.Values = append(b.Values, vb)
		}
		if len(b.Values) > 0 || len(b.Services) > 0 {
			store := loadBackups()
			if _, exists := store[t.ID]; !exists {
				store[t.ID] = b
				saveBackups(store)
			}
		}
	case "service":
		b := TweakBackup{ID: t.ID, Type: "service", AppliedAt: time.Now().Format(time.RFC3339)}
		for _, c := range splitCommands(t.Command) {
			sm := scConfigRe.FindStringSubmatch(c)
			if len(sm) != 3 {
				continue
			}
			svc := sm[1]
			if st, ok := readServiceStartType(svc); ok {
				b.Services = append(b.Services, ServiceBackup{Service: svc, StartType: st})
			}
		}
		if len(b.Services) > 0 {
			store := loadBackups()
			if _, exists := store[t.ID]; !exists {
				store[t.ID] = b
				saveBackups(store)
			}
		}
	}
}

// readRegValue returns the current type and data of a registry value.
func readRegValue(path, name string) (string, string, bool) {
	out, err := runRegQuery(path, name)
	if err != nil {
		return "", "", false
	}
	target := strings.ToLower(name)
	for _, line := range strings.Split(string(out), "\n") {
		trimmed := strings.TrimSpace(strings.TrimRight(line, "\r"))
		if trimmed == "" {
			continue
		}
		low := strings.ToLower(trimmed)
		isDefault := name == "(Default)" && strings.Contains(low, "(default)")
		if !isDefault && !strings.Contains(low, target) {
			continue
		}
		fields := strings.Fields(trimmed)
		typeIdx := -1
		for i, f := range fields {
			if strings.HasPrefix(strings.ToUpper(f), "REG_") {
				typeIdx = i
				break
			}
		}
		if typeIdx == -1 {
			continue
		}
		regType := fields[typeIdx]
		data := ""
		if idx := strings.Index(trimmed, regType); idx >= 0 {
			data = strings.TrimSpace(trimmed[idx+len(regType):])
		}
		return regType, data, true
	}
	return "", "", false
}

// readServiceStartType returns the current start type of a service mapped to an
// `sc config start=` token (auto/demand/disabled/boot/system).
func readServiceStartType(svc string) (string, bool) {
	out, err := powershellCommand(fmt.Sprintf(`(Get-Service -Name '%s').StartType`, svc)).CombinedOutput()
	if err != nil {
		return "", false
	}
	switch strings.TrimSpace(strings.ToLower(string(out))) {
	case "automatic", "auto":
		return "auto", true
	case "manual":
		return "demand", true
	case "disabled":
		return "disabled", true
	case "boot":
		return "boot", true
	case "system":
		return "system", true
	}
	return "", false
}

// RevertTweak restores the system state captured before a tweak was applied.
func RevertTweak(id string) TweakResult {
	store := loadBackups()
	b, ok := store[id]
	if !ok {
		return TweakResult{ID: id, Success: false, Message: "Brak kopii zapasowej — tweak nie był zastosowany przez program"}
	}

	var errs []string
	switch b.Type {
	case "registry":
		for _, v := range b.Values {
			var out []byte
			var err error
			switch {
			case v.Existed && v.Name == "(Default)":
				out, err = runReg("add", v.Path, "/ve", "/t", v.Type, "/d", v.Data, "/f")
			case v.Existed:
				out, err = runReg("add", v.Path, "/v", v.Name, "/t", v.Type, "/d", v.Data, "/f")
			case v.Name == "(Default)":
				out, err = runReg("delete", v.Path, "/ve", "/f")
			default:
				out, err = runReg("delete", v.Path, "/v", v.Name, "/f")
			}
			if err != nil && !isRegNotFound(out) {
				errs = append(errs, fmt.Sprintf("%s\\%s: %s (%s)", v.Path, v.Name, err.Error(), strings.TrimSpace(string(out))))
			}
		}
		// Registry-typed tweaks may have captured embedded service changes.
		for _, s := range b.Services {
			cmd := fmt.Sprintf(`sc config %s start= %s`, s.Service, s.StartType)
			if out, err := runShellCommand(cmd); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %s", cmd, strings.TrimSpace(string(out))))
			}
		}
	case "service":
		for _, s := range b.Services {
			cmd := fmt.Sprintf(`sc config %s start= %s`, s.Service, s.StartType)
			if out, err := runShellCommand(cmd); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %s", cmd, strings.TrimSpace(string(out))))
			}
		}
	default:
		return TweakResult{ID: id, Success: false, Message: "Tego typu nie można cofnąć automatycznie"}
	}

	if len(errs) > 0 {
		return TweakResult{ID: id, Success: false, Message: addAdminHint(strings.Join(errs, "; "))}
	}

	delete(store, id)
	saveBackups(store)
	return TweakResult{ID: id, Success: true, Message: "Cofnięto"}
}
