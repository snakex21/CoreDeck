package tweaks

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"launcher/internal/uimsg"
	"sync/atomic"

	"launcher/internal/jsonutil"
)

var (
	regAddRe    = regexp.MustCompile(`(?i)reg\s+add\s+"([^"]+)"\s+(.*)`)
	regDeleteRe = regexp.MustCompile(`(?i)reg\s+delete\s+"([^"]+)"\s+(.*)`)
	regValueRe  = regexp.MustCompile(`(?i)(?:^|\s)/v\s+([^\s]+)`)
	regDataRe   = regexp.MustCompile(`(?i)(?:^|\s)/d\s+([^\s]+)`)
	scConfigRe  = regexp.MustCompile(`(?i)sc\s+config\s+([^\s]+)\s+start=\s*([^\s;]+)`)
)

type Tweak struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Type        string   `json:"type"`
	Command     string   `json:"command"`
	Risk        string   `json:"risk"`
	Profiles    []string `json:"profiles"`
	OS          string   `json:"os,omitempty"`
	Device      string   `json:"device,omitempty"`
	GPU         string   `json:"gpu,omitempty"`
	Group       string   `json:"group,omitempty"`
	Situational bool     `json:"situational,omitempty"`
	Requires    []Check  `json:"requires,omitempty"`
	Checks      []Check  `json:"checks,omitempty"`
	Query       []Check  `json:"query,omitempty"`
}

type Check struct {
	Type       string   `json:"type"`
	Path       string   `json:"path,omitempty"`
	Name       string   `json:"name,omitempty"`
	Equals     string   `json:"equals,omitempty"`
	Command    string   `json:"command,omitempty"`
	Contains   string   `json:"contains,omitempty"`
	Service    string   `json:"service,omitempty"`
	Services   []string `json:"services,omitempty"`
	Startup    string   `json:"startup,omitempty"`
	Package    string   `json:"package,omitempty"`
	Capability string   `json:"capability,omitempty"`
	Feature    string   `json:"feature,omitempty"`
	File       string   `json:"file,omitempty"`
}

type TweakState struct {
	ID      string `json:"id"`
	Active  bool   `json:"active"`
	Known   bool   `json:"known"`
	Removed bool   `json:"removed"`
	Message string `json:"message,omitempty"`
	// Code is the UI message key (translations.js), Params its placeholders;
	// Message is the Polish fallback and log text (see internal/uimsg).
	Code       string            `json:"code,omitempty"`
	Params     map[string]string `json:"params,omitempty"`
	Revertible bool              `json:"revertible,omitempty"`
}

type TweaksData struct {
	Version    string            `json:"version"`
	Categories map[string]string `json:"categories"`
	Tweaks     []Tweak           `json:"tweaks"`
}

type TweakResult struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	// Code / Params: see TweakState.
	Code   string            `json:"code,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

// newResult builds a TweakResult carrying a UI message.
func newResult(id, name string, success bool, m uimsg.Msg) TweakResult {
	return TweakResult{ID: id, Name: name, Success: success, Message: m.Text, Code: m.Code, Params: m.Params}
}

// setMsg sets the state's UI message.
func (s *TweakState) setMsg(m uimsg.Msg) {
	s.Message, s.Code, s.Params = m.Text, m.Code, m.Params
}

func LoadTweaks(webDir string) (*TweaksData, error) {
	jsonPath := filepath.Join(webDir, "data", "tweaks.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read tweaks.json: %w", err)
	}
	var td TweaksData
	if err := jsonutil.Unmarshal(data, &td); err != nil {
		return nil, fmt.Errorf("cannot parse tweaks.json: %w", err)
	}
	populateInferredChecks(&td)
	return &td, nil
}

func LoadFromFS(fsys fs.FS) (*TweaksData, error) {
	data, err := fs.ReadFile(fsys, "web/data/tweaks.json")
	if err != nil {
		return nil, fmt.Errorf("cannot read tweaks.json from FS: %w", err)
	}
	var td TweaksData
	if err := jsonutil.Unmarshal(data, &td); err != nil {
		return nil, fmt.Errorf("cannot parse tweaks.json: %w", err)
	}
	populateInferredChecks(&td)
	return &td, nil
}

func populateInferredChecks(td *TweaksData) {
	for i := range td.Tweaks {
		if len(td.Tweaks[i].Checks) == 0 && len(td.Tweaks[i].Query) > 0 {
			td.Tweaks[i].Checks = td.Tweaks[i].Query
		}
		if len(td.Tweaks[i].Checks) == 0 {
			td.Tweaks[i].Checks = inferChecks(td.Tweaks[i])
		}
	}
}

func inferChecks(t Tweak) []Check {
	if checks := explicitChecks(t.ID); len(checks) > 0 {
		return checks
	}

	var checks []Check
	commands := splitCommands(t.Command)
	for _, command := range commands {
		c := strings.TrimSpace(command)
		if c == "" {
			continue
		}
		lower := strings.ToLower(c)

		if strings.Contains(lower, "powercfg -setactive") {
			parts := strings.Fields(c)
			guid := parts[len(parts)-1]
			checks = append(checks, Check{Type: "command", Command: "powercfg /getactivescheme", Contains: guid})
			continue
		}

		if check, ok := inferPowercfgChangeCheck(c); ok {
			checks = append(checks, check)
			continue
		}

		if strings.Contains(lower, "powercfg -hibernate off") || strings.Contains(lower, "powercfg /hibernate off") {
			checks = append(checks, Check{Type: "registry", Path: `HKLM\SYSTEM\CurrentControlSet\Control\Power`, Name: "HibernateEnabled", Equals: "0"})
			continue
		}

		if strings.Contains(lower, "disable-mmagent") && strings.Contains(lower, "memorycompression") {
			checks = append(checks, Check{Type: "command", Command: `powershell -NoProfile -Command "(Get-MMAgent).MemoryCompression"`, Contains: "False"})
			continue
		}

		if m := scConfigRe.FindStringSubmatch(c); len(m) == 3 {
			startup := strings.ToLower(strings.TrimSpace(m[2]))
			if startup == "demand" {
				startup = "manual"
			}
			checks = append(checks, Check{Type: "service", Service: m[1], Startup: startup})
			continue
		}

		if m := regAddRe.FindStringSubmatch(c); len(m) == 3 {
			check := Check{Type: "registry", Path: m[1]}
			rest := m[2]
			if strings.Contains(strings.ToLower(rest), "/ve") {
				check.Name = "(Default)"
			} else if vm := regValueRe.FindStringSubmatch(rest); len(vm) == 2 {
				check.Name = vm[1]
			}
			if dm := regDataRe.FindStringSubmatch(rest); len(dm) == 2 {
				check.Equals = strings.Trim(dm[1], `"`)
			}
			if check.Name != "" {
				checks = append(checks, check)
			}
		}
	}
	return checks
}

func inferPowercfgChangeCheck(command string) (Check, bool) {
	parts := strings.Fields(strings.ToLower(command))
	if len(parts) < 4 || parts[0] != "powercfg" || parts[1] != "-change" {
		return Check{}, false
	}

	option := parts[2]
	value := parts[3]

	setting := ""
	switch {
	case strings.HasPrefix(option, "-monitor-timeout"):
		setting = "SUB_VIDEO VIDEOIDLE"
	case strings.HasPrefix(option, "-standby-timeout"):
		setting = "SUB_SLEEP STANDBYIDLE"
	case strings.HasPrefix(option, "-hibernate-timeout"):
		setting = "SUB_SLEEP HIBERNATEIDLE"
	default:
		return Check{}, false
	}

	contains := fmt.Sprintf("0x%08x", parsePowercfgMinutes(value))
	return Check{Type: "command", Command: "powercfg /query SCHEME_CURRENT " + setting, Contains: contains}, true
}

func explicitChecks(id string) []Check {
	switch id {
	case "ultimate_performance":
		return []Check{{Type: "command", Command: "powercfg /getactivescheme", Contains: "e9a42b02-d5df-448d-aa00-03f14749eb61"}}
	case "never_sleep":
		return []Check{
			{Type: "command", Command: "powercfg /query SCHEME_CURRENT SUB_VIDEO VIDEOIDLE", Contains: "0x00000000"},
			{Type: "command", Command: "powercfg /query SCHEME_CURRENT SUB_SLEEP STANDBYIDLE", Contains: "0x00000000"},
		}
	case "laptop_power":
		return []Check{{Type: "command", Command: "powercfg /getactivescheme", Contains: "381b4222-f694-41f0-9685-ff5bb260df2e"}}
	case "disable_memory_compression":
		return []Check{{Type: "command", Command: "powershell -NoProfile -Command \"(Get-MMAgent).MemoryCompression\"", Contains: "False"}}
	}
	return nil
}

func parsePowercfgMinutes(value string) int {
	var result int
	fmt.Sscanf(value, "%d", &result)
	return result * 60
}

func GetTweaksForProfile(td *TweaksData, profileName string) []Tweak {
	var result []Tweak
	for _, t := range td.Tweaks {
		for _, p := range t.Profiles {
			if p == profileName {
				result = append(result, t)
				break
			}
		}
	}
	return result
}

func CheckTweakState(t Tweak) TweakState {
	state := TweakState{ID: t.ID, Known: true}

	// Check if required components exist
	if len(t.Requires) > 0 {
		for _, req := range t.Requires {
			if !checkRequirement(req) {
				state.Active = false
				state.Removed = true
				state.Known = true
				state.setMsg(uimsg.New("msg_tws_component_removed", "Komponent usunięty z systemu"))
				return state
			}
		}
	}

	if len(t.Checks) == 0 {
		state.Known = false
		state.setMsg(uimsg.New("msg_tws_no_check_rule", "Brak reguły sprawdzania"))
		return state
	}

	for _, check := range t.Checks {
		ok, msg := RunCheck(check)
		if !ok {
			state.Active = false
			state.setMsg(msg)
			return state
		}
	}

	state.Active = true
	state.setMsg(uimsg.New("msg_tws_active", "Aktywny"))
	return state
}

func checkRequirement(req Check) bool {
	switch req.Type {
	case "package":
		if req.Package == "" {
			return true
		}
		if installed, known := cachedPackageInstalled(req.Package); known {
			return installed
		}
		return true // assume present when the shared AppX scan failed

	case "capability":
		if req.Capability == "" {
			return true
		}
		cmd := powershellCommand(fmt.Sprintf(`$c = Get-WindowsCapability -Name '%s' -Online -ErrorAction SilentlyContinue; if ($c -and $c.State -ne 'Removed') { 'installed' } else { 'removed' }`, req.Capability))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return true
		}
		return strings.TrimSpace(string(output)) == "installed"

	case "feature":
		if req.Feature == "" {
			return true
		}
		cmd := powershellCommand(fmt.Sprintf(`$f = Get-WindowsOptionalFeature -FeatureName '%s' -Online -ErrorAction SilentlyContinue; if ($f -and $f.State -ne 'Disabled' -and $f.State -ne 'DisabledWithPayloadRemoved') { 'installed' } else { 'removed' }`, req.Feature))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return true
		}
		return strings.TrimSpace(string(output)) == "installed"

	case "file":
		if req.File == "" {
			return true
		}
		_, err := os.Stat(req.File)
		return err == nil

	case "service":
		services := req.Services
		if len(services) == 0 && req.Service != "" {
			services = []string{req.Service}
		}
		for _, service := range services {
			if _, exists, known := cachedServiceStartType(service); known && !exists {
				return false
			}
		}
		return true

	default:
		return true
	}
}

func CheckTweaksStates(tweaks []Tweak) []TweakState {
	return CheckTweaksStatesWithProgress(tweaks, nil)
}

func CheckTweaksStatesWithProgress(tweaks []Tweak, progress func(completed, total int)) []TweakState {
	refreshStateCaches()
	backups := loadBackups()
	states := make([]TweakState, len(tweaks))

	// Parallel state checking — each tweak check is independent.
	sem := make(chan struct{}, 8) // max 8 concurrent checks
	var wg sync.WaitGroup
	var completed atomic.Int64

	for i, tw := range tweaks {
		wg.Add(1)
		go func(i int, tw Tweak) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			s := CheckTweakState(tw)
			if IsRevertible(tw) {
				if _, ok := backups[tw.ID]; ok {
					s.Revertible = true
				}
			}
			states[i] = s
			if progress != nil {
				progress(int(completed.Add(1)), len(tweaks))
			}
		}(i, tw)
	}
	wg.Wait()
	return states
}

// RunCheck evaluates one check rule; the message explains the result.
func RunCheck(check Check) (bool, uimsg.Msg) {
	switch check.Type {
	case "registry":
		return checkRegistry(check)
	case "service":
		return checkService(check)
	case "command":
		return checkCommand(check)
	case "package":
		return checkPackage(check)
	default:
		return false, uimsg.New("msg_tws_unknown_check", "Nieznany typ sprawdzania: "+check.Type, "type", check.Type)
	}
}

var (
	stateCacheMu      sync.RWMutex
	appxCache         map[string]struct{}
	appxCacheReady    bool
	serviceCache      map[string]string
	serviceCacheReady bool
)

func refreshStateCaches() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		refreshAppxCache()
	}()
	go func() {
		defer wg.Done()
		refreshServiceCache()
	}()
	wg.Wait()
}

// refreshAppxCache loads all installed AppX package names once (case-insensitive
// set) so per-package checks don't each spawn PowerShell.
func refreshAppxCache() {
	out, err := powershellCommand(`Get-AppxPackage -AllUsers | Select-Object -ExpandProperty Name`).CombinedOutput()
	if err != nil {
		stateCacheMu.Lock()
		appxCache = nil
		appxCacheReady = false
		stateCacheMu.Unlock()
		return
	}
	m := map[string]struct{}{}
	for _, line := range strings.Split(string(out), "\n") {
		name := strings.TrimSpace(strings.TrimRight(line, "\r"))
		if name != "" {
			m[strings.ToLower(name)] = struct{}{}
		}
	}
	stateCacheMu.Lock()
	appxCache = m
	appxCacheReady = true
	stateCacheMu.Unlock()
}

func refreshServiceCache() {
	out, err := powershellCommand(`Get-Service | ForEach-Object { [string]$_.Name + [char]9 + [string]$_.StartType }`).CombinedOutput()
	if err != nil {
		stateCacheMu.Lock()
		serviceCache = nil
		serviceCacheReady = false
		stateCacheMu.Unlock()
		return
	}
	m := map[string]string{}
	for _, line := range strings.Split(string(out), "\n") {
		parts := strings.SplitN(strings.TrimSpace(strings.TrimRight(line, "\r")), "\t", 2)
		if len(parts) == 2 && parts[0] != "" {
			m[strings.ToLower(parts[0])] = strings.ToLower(strings.TrimSpace(parts[1]))
		}
	}
	stateCacheMu.Lock()
	serviceCache = m
	serviceCacheReady = true
	stateCacheMu.Unlock()
}

func cachedPackageInstalled(name string) (installed bool, known bool) {
	stateCacheMu.RLock()
	defer stateCacheMu.RUnlock()
	if !appxCacheReady {
		return false, false
	}
	target := strings.ToLower(strings.TrimSuffix(name, "*"))
	for installedName := range appxCache {
		if strings.HasPrefix(installedName, target) {
			return true, true
		}
	}
	return false, true
}

func cachedServiceStartType(name string) (startType string, exists bool, known bool) {
	stateCacheMu.RLock()
	defer stateCacheMu.RUnlock()
	if !serviceCacheReady {
		return "", false, false
	}
	startType, exists = serviceCache[strings.ToLower(name)]
	return startType, exists, true
}

// checkPackage reports a debloat tweak as ACTIVE when the appx package is NOT
// installed (i.e. it has been removed).
func checkPackage(check Check) (bool, uimsg.Msg) {
	name := check.Package
	if name == "" {
		name = check.Name
	}
	if name == "" {
		return false, uimsg.New("msg_tws_no_package_name", "Brak nazwy pakietu")
	}
	if installed, known := cachedPackageInstalled(name); known {
		if installed {
			return false, uimsg.New("msg_tws_installed", "Zainstalowany")
		}
		return true, uimsg.New("msg_tws_removed", "Usunięty")
	}
	return false, uimsg.New("msg_tws_packages_unreadable", "Nie można odczytać listy pakietów")
}

// checkRegistry parses reg query output properly, line by line.
func checkRegistry(check Check) (bool, uimsg.Msg) {
	if check.Path == "" || check.Name == "" {
		return false, uimsg.New("msg_tws_registry_rule_incomplete", "Niepełna reguła registry")
	}

	output, err := runRegQuery(check.Path, check.Name)
	if err != nil {
		entry := check.Path + "\\" + check.Name
		return false, uimsg.New("msg_tws_registry_missing", "Brak wpisu: "+entry, "entry", entry)
	}

	outputStr := string(output)

	// Parse: find the line containing the value name
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// For (Default), look for lines with "(Default)"
		isDefaultLine := check.Name == "(Default)" && strings.Contains(strings.ToLower(line), "(default)")

		// For named values, check if line contains the value name
		isNamedLine := false
		if !isDefaultLine {
			lowerLine := strings.ToLower(line)
			lowerName := strings.ToLower(check.Name)
			if strings.Contains(lowerLine, lowerName) {
				// Verify it's actually the value line by checking for REG_ type
				if strings.Contains(lowerLine, "reg_sz") || strings.Contains(lowerLine, "reg_dword") ||
					strings.Contains(lowerLine, "reg_qword") || strings.Contains(lowerLine, "reg_binary") ||
					strings.Contains(lowerLine, "reg_expand_sz") || strings.Contains(lowerLine, "reg_multi_sz") {
					isNamedLine = true
				}
			}
		}

		if !isDefaultLine && !isNamedLine {
			continue
		}

		// Found the value line. Parse: "Name    TYPE    Value"
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		// Last part is the data value (may be empty for (Default) REG_SZ)
		actualValue := ""
		if len(parts) >= 3 {
			actualValue = parts[len(parts)-1]
		}

		// If no expected value, just checking existence
		if check.Equals == "" {
			return true, uimsg.New("msg_tws_entry_exists", "Wpis istnieje")
		}

		// Compare
		if registryValueMatches(actualValue, check.Equals) {
			return true, uimsg.New("msg_tws_matches", "Zgodny")
		}
		return false, uimsg.New("msg_tws_value_mismatch", fmt.Sprintf("Oczekiwano %s, jest %s", check.Equals, actualValue),
			"expected", check.Equals, "actual", actualValue)
	}

	return false, uimsg.New("msg_tws_value_not_found", "Nie znaleziono wartości w output")
}

// registryValueMatches compares registry values handling DWORD formats
func registryValueMatches(actual, expected string) bool {
	actual = strings.TrimSpace(strings.ToLower(actual))
	expected = strings.TrimSpace(strings.ToLower(expected))

	// Direct match
	if actual == expected {
		return true
	}

	// DWORD comparison: normalize 0x00000000 == 0x0 == 0
	actualNum := parseHexOrDec(actual)
	expectedNum := parseHexOrDec(expected)
	if actualNum != nil && expectedNum != nil {
		return *actualNum == *expectedNum
	}

	return actual == expected
}

func parseHexOrDec(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var val int
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		n, err := fmt.Sscanf(s, "0x%x", &val)
		if n == 1 && err == nil {
			return &val
		}
	} else {
		n, err := fmt.Sscanf(s, "%d", &val)
		if n == 1 && err == nil {
			return &val
		}
	}
	return nil
}

// checkService uses PowerShell to get service start type (locale-independent).
func checkService(check Check) (bool, uimsg.Msg) {
	services := check.Services
	if len(services) == 0 && check.Service != "" {
		services = []string{check.Service}
	}
	if len(services) == 0 {
		return false, uimsg.New("msg_tws_no_service_name", "Brak nazwy usługi")
	}

	expected := strings.ToLower(check.Startup)
	if expected == "demand" {
		expected = "manual"
	}

	for _, svc := range services {
		if actual, exists, known := cachedServiceStartType(svc); known {
			if !exists {
				return false, serviceUnreadable(svc)
			}
			if !serviceStartTypeMatches(actual, expected) {
				return false, serviceMismatch(svc, actual, expected)
			}
			continue
		}

		cmd := powershellCommand(fmt.Sprintf(`(Get-Service -Name '%s').StartType`, svc))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return false, serviceUnreadable(svc)
		}

		actual := strings.TrimSpace(strings.ToLower(string(output)))

		if !serviceStartTypeMatches(actual, expected) {
			return false, serviceMismatch(svc, actual, expected)
		}
	}
	return true, uimsg.New("msg_tws_services_match", "Usługi zgodne")
}

func serviceUnreadable(svc string) uimsg.Msg {
	return uimsg.New("msg_tws_service_unreadable", "Nie można odczytać usługi: "+svc, "service", svc)
}

func serviceMismatch(svc, actual, expected string) uimsg.Msg {
	return uimsg.New("msg_tws_service_mismatch", fmt.Sprintf("Usługa %s: start=%s, oczekiwano %s", svc, actual, expected),
		"service", svc, "actual", actual, "expected", expected)
}

func serviceStartTypeMatches(actual, expected string) bool {
	actual = strings.ToLower(strings.TrimSpace(actual))
	expected = strings.ToLower(strings.TrimSpace(expected))
	if expected == "demand" {
		expected = "manual"
	}
	if expected == "auto" {
		expected = "automatic"
	}
	return actual == expected || (expected == "automatic" && actual == "auto")
}

// checkCommand runs a command and checks output for expected string.
func checkCommand(check Check) (bool, uimsg.Msg) {
	if check.Command == "" {
		return false, uimsg.New("msg_tws_no_command", "Brak komendy sprawdzającej")
	}

	var output []byte
	var err error

	// For PowerShell commands, use powershell directly (not through cmd /C)
	lowerCmd := strings.ToLower(strings.TrimSpace(check.Command))
	if strings.HasPrefix(lowerCmd, "powershell") {
		// Extract the -Command part
		idx := strings.Index(strings.ToLower(check.Command), "-command")
		if idx >= 0 {
			psCmd := strings.TrimSpace(check.Command[idx+8:])
			psCmd = strings.Trim(psCmd, `"`)
			cmd := powershellCommand(psCmd)
			output, err = cmd.CombinedOutput()
		} else {
			output, err = runShellCommand(check.Command)
		}
	} else {
		output, err = runShellCommand(check.Command)
	}

	if err != nil {
		return false, uimsg.New("msg_tws_command_failed", "Komenda sprawdzająca nie powiodła się: "+err.Error(), "error", err.Error())
	}
	out := strings.ToLower(string(output))

	if check.Contains != "" {
		if !strings.Contains(out, strings.ToLower(check.Contains)) {
			got := strings.TrimSpace(string(output))
			return false, uimsg.New("msg_tws_output_missing", fmt.Sprintf("Wynik nie zawiera '%s', output: %s", check.Contains, got),
				"expected", check.Contains, "output", got)
		}
	}
	if check.Equals != "" {
		if strings.TrimSpace(out) != strings.ToLower(strings.TrimSpace(check.Equals)) {
			got := strings.TrimSpace(string(output))
			return false, uimsg.New("msg_tws_output_mismatch", fmt.Sprintf("Wynik '%s' != '%s'", got, check.Equals),
				"actual", got, "expected", check.Equals)
		}
	}
	return true, uimsg.New("msg_tws_matches", "Zgodny")
}

type TweakProgress struct {
	Current int           `json:"current"`
	Total   int           `json:"total"`
	Name    string        `json:"name"`
	Results []TweakResult `json:"results"`
}

func ApplyTweaks(tweaks []Tweak, progressCb func(TweakProgress), cancelCheck func() bool) []TweakResult {
	var results []TweakResult
	total := len(tweaks)

	for i, t := range tweaks {
		if cancelCheck != nil && cancelCheck() {
			break
		}

		state := CheckTweakState(t)
		if state.Known && state.Active {
			results = append(results, newResult(t.ID, t.Name, true, uimsg.New("msg_tw_already_active", "Już aktywny — pominięto")))
		} else {
			result := ApplyTweak(t)
			results = append(results, result)
		}

		if progressCb != nil {
			resultsCopy := make([]TweakResult, len(results))
			copy(resultsCopy, results)
			progressCb(TweakProgress{
				Current: i + 1,
				Total:   total,
				Name:    t.Name,
				Results: resultsCopy,
			})
		}
	}

	return results
}

func ApplyTweak(t Tweak) TweakResult {
	captureBackup(t)
	switch t.Type {
	case "registry":
		commands := splitCommands(t.Command)
		var errors []string
		for _, c := range commands {
			c = strings.TrimSpace(c)
			if c == "" {
				continue
			}
			// A registry-typed tweak may embed non-reg sub-commands (e.g. sc,
			// net, netsh, powercfg). Route each by its leading token: reg/reg.exe
			// go to reg.exe directly, everything else runs via cmd.exe.
			if isRegCommand(c) {
				output, err := runRegCommand(c)
				if err != nil {
					if isRegNotFound(output) && strings.Contains(strings.ToLower(c), "reg delete") {
						continue
					}
					errors = append(errors, fmt.Sprintf("%s: %s (%s)", c, err.Error(), strings.TrimSpace(string(output))))
				}
				continue
			}
			output, err := runShellCommand(c)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %s (%s)", c, err.Error(), strings.TrimSpace(string(output))))
			}
		}
		if len(errors) > 0 {
			return newResult(t.ID, t.Name, false, failureMessage(strings.Join(errors, "; ")))
		}
		return newResult(t.ID, t.Name, true, applied)

	case "powershell":
		cmd := powershellCommand(t.Command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return newResult(t.ID, t.Name, false, failureMessage(err.Error()+": "+strings.TrimSpace(string(output))))
		}
		return newResult(t.ID, t.Name, true, applied)

	case "powercfg":
		commands := splitCommands(t.Command)
		var errors []string
		for _, c := range commands {
			c = strings.TrimSpace(c)
			if c == "" {
				continue
			}
			output, err := runShellCommand(c)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %s", c, strings.TrimSpace(string(output))))
			}
		}
		if len(errors) > 0 {
			return newResult(t.ID, t.Name, false, failureMessage(strings.Join(errors, "; ")))
		}
		return newResult(t.ID, t.Name, true, applied)

	case "service":
		commands := splitCommands(t.Command)
		var errors []string
		for _, c := range commands {
			c = strings.TrimSpace(c)
			if c == "" {
				continue
			}
			output, err := runShellCommand(c)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %s", c, strings.TrimSpace(string(output))))
			}
		}
		if len(errors) > 0 {
			return newResult(t.ID, t.Name, false, failureMessage(strings.Join(errors, "; ")))
		}
		return newResult(t.ID, t.Name, true, applied)

	case "appx":
		var errs []string
		for _, name := range splitCommands(t.Command) {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			ps := fmt.Sprintf(`Get-AppxPackage -AllUsers -Name '%s' | Remove-AppxPackage -AllUsers -ErrorAction SilentlyContinue; Get-AppxProvisionedPackage -Online | Where-Object { $_.DisplayName -like '%s' } | Remove-AppxProvisionedPackage -Online -AllUsers -ErrorAction SilentlyContinue`, name, name)
			output, err := powershellCommand(ps).CombinedOutput()
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s: %s (%s)", name, err.Error(), strings.TrimSpace(string(output))))
			}
		}
		if len(errs) > 0 {
			return newResult(t.ID, t.Name, false, failureMessage(strings.Join(errs, "; ")))
		}
		return newResult(t.ID, t.Name, true, uimsg.New("msg_tw_removed", "Usunięto"))

	default:
		return newResult(t.ID, t.Name, false, uimsg.New("msg_tw_unknown_type", "Nieznany typ: "+t.Type, "type", t.Type))
	}
}

var applied = uimsg.New("msg_tw_applied", "Zastosowano")

// failureMessage wraps the raw command errors; access errors get a hint to
// run CoreDeck as administrator.
func failureMessage(details string) uimsg.Msg {
	lower := strings.ToLower(details)
	if strings.Contains(lower, "access is denied") || strings.Contains(lower, "odmowa dostępu") || strings.Contains(lower, "denied") {
		return uimsg.New("msg_tw_failed_admin", details+" — uruchom launcher jako administrator", "error", details)
	}
	return uimsg.New("msg_tw_failed", details, "error", details)
}

// psExe returns the PowerShell executable to use, preferring Windows PowerShell
// (present on stock Win10/11) and falling back to PowerShell 7 (pwsh).
func psExe() string {
	if p, err := exec.LookPath("powershell"); err == nil {
		return p
	}
	if p, err := exec.LookPath("pwsh"); err == nil {
		return p
	}
	return "powershell"
}

// powershellCommand builds a PowerShell invocation with consistent flags
// (-NoProfile and -ExecutionPolicy Bypass) so behavior matches across all call
// sites regardless of machine execution policy.
func powershellCommand(script string) *exec.Cmd {
	return hidden(exec.Command(psExe(), "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script))
}

// splitCommands splits a command string on ';' separators while respecting
// double-quoted segments, so a ';' inside a quoted value (e.g. a reg data
// value) is not mistaken for a command separator.
func splitCommands(s string) []string {
	var parts []string
	var b strings.Builder
	inQuote := false
	for _, r := range s {
		switch r {
		case '"':
			inQuote = !inQuote
			b.WriteRune(r)
		case ';':
			if inQuote {
				b.WriteRune(r)
			} else {
				parts = append(parts, b.String())
				b.Reset()
			}
		default:
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 {
		parts = append(parts, b.String())
	}
	return parts
}

func runShellCommand(command string) ([]byte, error) {
	return hidden(exec.Command("cmd.exe", "/C", command)).CombinedOutput()
}

// tokenizeArgs splits a command string into arguments, honoring double quotes
// (quoted spans are kept together and the quotes removed) so paths/values that
// contain spaces (e.g. "Windows Search") survive intact.
func tokenizeArgs(s string) []string {
	var args []string
	var b strings.Builder
	inQuote := false
	hasToken := false
	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
			hasToken = true
		case (r == ' ' || r == '\t') && !inQuote:
			if hasToken {
				args = append(args, b.String())
				b.Reset()
				hasToken = false
			}
		default:
			b.WriteRune(r)
			hasToken = true
		}
	}
	if hasToken {
		args = append(args, b.String())
	}
	return args
}

// runReg runs reg.exe directly with the given arguments (no cmd.exe), which
// avoids the double quoting/escaping that breaks keys or values containing
// spaces.
func runReg(args ...string) ([]byte, error) {
	return hidden(exec.Command("reg.exe", args...)).CombinedOutput()
}

// isRegCommand reports whether a sub-command's leading token is reg/reg.exe so
// it should be routed to reg.exe rather than the shell.
func isRegCommand(command string) bool {
	tokens := tokenizeArgs(command)
	if len(tokens) == 0 {
		return false
	}
	return strings.EqualFold(tokens[0], "reg") || strings.EqualFold(tokens[0], "reg.exe")
}

// runRegCommand parses a "reg add/delete/query ..." command string and runs
// reg.exe directly with properly separated arguments.
func runRegCommand(command string) ([]byte, error) {
	tokens := tokenizeArgs(command)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("pusta komenda reg")
	}
	if strings.EqualFold(tokens[0], "reg") || strings.EqualFold(tokens[0], "reg.exe") {
		tokens = tokens[1:]
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("niekompletna komenda reg")
	}
	return runReg(tokens...)
}

// isRegNotFound reports whether reg.exe output indicates the key/value was not
// found (so a reg delete that has nothing to delete can be treated as success).
func isRegNotFound(output []byte) bool {
	low := strings.ToLower(string(output))
	return strings.Contains(low, "unable to find") ||
		strings.Contains(low, "cannot find") ||
		strings.Contains(low, "the system was unable to find") ||
		strings.Contains(low, "nie może odnaleźć") ||
		strings.Contains(low, "nie moze odnalezc")
}

// runRegQuery runs reg query directly without cmd.exe wrapper
func runRegQuery(path string, name string) ([]byte, error) {
	if name == "(Default)" {
		return hidden(exec.Command("reg.exe", "query", path, "/ve")).CombinedOutput()
	}
	return hidden(exec.Command("reg.exe", "query", path, "/v", name)).CombinedOutput()
}
