package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"launcher/internal/drivers"
	"launcher/internal/programs"
	"launcher/internal/system"
)

const (
	toolName          = "disk-helper"
	schemaVersion     = "1.0.0"
	maxContextEntries = 240
	maxContextWalk    = 5000
	maxFindMatches    = 120
	maxFindScanned    = 120000
	maxCatalogFolders = 300
)

type envelope struct {
	Tool          string      `json:"tool"`
	SchemaVersion string      `json:"schemaVersion"`
	Command       commandMeta `json:"command"`
	Root          string      `json:"root"`
	OK            bool        `json:"ok"`
	Warnings      []string    `json:"warnings"`
	Errors        []string    `json:"errors"`
	Truncated     bool        `json:"truncated"`
	Data          any         `json:"data"`
}

type commandMeta struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

type options struct {
	Root         string
	LauncherDir  string
	ProgramsJSON string
	DriversJSON  string
}

type fileEntry struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	RelPath  string    `json:"relPath,omitempty"`
	Kind     string    `json:"kind"`
	Ext      string    `json:"ext,omitempty"`
	Size     int64     `json:"size,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
}

type folderCandidate struct {
	Path     string   `json:"path"`
	RelPath  string   `json:"relPath,omitempty"`
	Category string   `json:"category,omitempty"`
	Source   string   `json:"source"`
	Score    int      `json:"score"`
	Evidence []string `json:"evidence,omitempty"`
}

type programMatch struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	InstallType string   `json:"installType"`
	FolderPath  string   `json:"folderPath,omitempty"`
	Path        string   `json:"path,omitempty"`
	Score       int      `json:"score"`
	Evidence    []string `json:"evidence,omitempty"`
}

type folderRecord struct {
	Path     string
	Category string
	Source   string
	Tokens   []string
}

type validationIssue struct {
	Name     string   `json:"name"`
	Category string   `json:"category,omitempty"`
	Field    string   `json:"field"`
	Path     string   `json:"path,omitempty"`
	Message  string   `json:"message"`
	Evidence []string `json:"evidence,omitempty"`
}

func main() {
	opts, cmd, args, warnings, err := parseArgs(os.Args[1:])
	if err != nil {
		writeEnvelope(envelopeFor(cmd, args, opts, warnings, []string{err.Error()}, true, false, usageData()))
		os.Exit(2)
	}

	data, moreWarnings, truncated, err := run(cmd, args, opts)
	warnings = append(warnings, moreWarnings...)
	var errs []string
	if err != nil {
		errs = append(errs, err.Error())
	}
	writeEnvelope(envelopeFor(cmd, args, opts, warnings, errs, len(errs) == 0, truncated, data))
	if len(errs) > 0 {
		os.Exit(1)
	}
}

func parseArgs(args []string) (options, string, []string, []string, error) {
	var opts options
	fs := flag.NewFlagSet(toolName, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&opts.Root, "root", "", "drive/root to scan (default: drive of --launcher)")
	fs.StringVar(&opts.LauncherDir, "launcher", "", "launcher dir holding CoreDeck.exe (default: detected from this exe)")
	fs.StringVar(&opts.ProgramsJSON, "programs", "", "programs.json path")
	fs.StringVar(&opts.DriversJSON, "drivers", "", "drivers.json path")
	if err := fs.Parse(args); err != nil {
		return opts, "", nil, nil, err
	}
	remaining := fs.Args()
	if len(remaining) == 0 {
		return opts, "", nil, nil, errors.New("missing command")
	}
	cmd := strings.ToLower(remaining[0])
	cmdArgs := remaining[1:]
	if cmd == "help" || cmd == "--help" || cmd == "-h" {
		return opts, "help", cmdArgs, nil, nil
	}
	if cmd != "context" && cmd != "find" && cmd != "validate" && cmd != "folders" {
		return opts, cmd, cmdArgs, nil, fmt.Errorf("unknown command %q", cmd)
	}
	warnings := normalizeOptions(&opts)
	return opts, cmd, cmdArgs, warnings, nil
}

func normalizeOptions(opts *options) []string {
	var warnings []string
	if strings.TrimSpace(opts.LauncherDir) == "" {
		opts.LauncherDir = detectLauncherDir()
	} else if abs, err := filepath.Abs(opts.LauncherDir); err == nil {
		opts.LauncherDir = abs
	}
	if strings.TrimSpace(opts.Root) == "" {
		drive := system.DriveOf(opts.LauncherDir)
		opts.Root = drive + "\\"
	} else if abs, err := filepath.Abs(opts.Root); err == nil {
		opts.Root = abs
	}
	if opts.ProgramsJSON == "" {
		opts.ProgramsJSON = filepath.Join(opts.LauncherDir, "data", "programs.json")
	}
	if opts.DriversJSON == "" {
		opts.DriversJSON = filepath.Join(opts.LauncherDir, "data", "drivers.json")
	}
	if _, err := os.Stat(opts.Root); err != nil {
		warnings = append(warnings, "root path is not currently accessible: "+opts.Root)
	}
	if _, err := os.Stat(opts.ProgramsJSON); err != nil {
		warnings = append(warnings, "programs.json not found: "+opts.ProgramsJSON)
	}
	if _, err := os.Stat(opts.DriversJSON); err != nil {
		warnings = append(warnings, "drivers.json not found: "+opts.DriversJSON)
	}
	return warnings
}

func detectLauncherDir() string {
	if env := strings.TrimSpace(os.Getenv("LAUNCHER_DIR")); env != "" {
		if abs, err := filepath.Abs(env); err == nil {
			return abs
		}
		return env
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates := []string{dir, filepath.Dir(dir), filepath.Dir(filepath.Dir(dir))}
		for _, c := range candidates {
			if fileExists(filepath.Join(c, "data", "programs.json")) {
				return c
			}
		}
		if strings.EqualFold(filepath.Base(dir), "tools") {
			return filepath.Dir(dir)
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates := []string{cwd, filepath.Join(cwd, "..", "bin"), filepath.Join(cwd, "..", "..", "bin")}
		for _, c := range candidates {
			c = filepath.Clean(c)
			if fileExists(filepath.Join(c, "data", "programs.json")) {
				return c
			}
		}
		return cwd
	}
	return "."
}

func run(cmd string, args []string, opts options) (any, []string, bool, error) {
	switch cmd {
	case "help":
		return usageData(), nil, false, nil
	case "context":
		if len(args) == 0 {
			return nil, nil, false, errors.New("context requires a path argument")
		}
		return commandContext(opts, strings.Join(args, " "))
	case "find":
		if len(args) == 0 {
			return nil, nil, false, errors.New("find requires a query argument")
		}
		return commandFind(opts, strings.Join(args, " "))
	case "validate":
		return commandValidate(opts)
	case "folders":
		return commandFolders(opts)
	default:
		return usageData(), nil, false, fmt.Errorf("unknown command %q", cmd)
	}
}

func commandContext(opts options, rawPath string) (any, []string, bool, error) {
	target := resolveInputPath(opts.Root, rawPath)
	info, err := os.Stat(target)
	if err != nil {
		return map[string]any{"path": target, "input": rawPath}, nil, false, err
	}

	entries, entryWarnings, entriesTruncated := listEntries(target, opts.Root)
	stats, evidence, statsTruncated := scanStats(target)
	query := filepath.Base(target)
	folders, _ := collectFolderRecords(opts)
	programsData, _ := loadProgramsFromPath(opts.ProgramsJSON)

	data := map[string]any{
		"input":                  rawPath,
		"path":                   target,
		"exists":                 true,
		"kind":                   kindOf(info),
		"size":                   info.Size(),
		"modified":               info.ModTime(),
		"stats":                  stats,
		"entries":                entries,
		"evidence":               evidence,
		"folderCandidates":       rankFolders(folders, query+" "+strings.Join(entryNames(entries), " "), opts.Root, 20),
		"existingProgramMatches": rankPrograms(programsData, query+" "+strings.Join(entryNames(entries), " "), opts.LauncherDir, 20),
		"suggestedQuestions":     suggestedQuestions(evidence, rankFolders(folders, query, opts.Root, 5), rankPrograms(programsData, query, opts.LauncherDir, 5)),
	}
	warnings := entryWarnings
	return data, warnings, entriesTruncated || statsTruncated, nil
}

func commandFind(opts options, query string) (any, []string, bool, error) {
	qTokens := tokenize(query)
	if len(qTokens) == 0 {
		return nil, nil, false, errors.New("empty query")
	}
	var matches []fileEntry
	scanned := 0
	truncated := false
	err := filepath.WalkDir(opts.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() && shouldSkipDir(name) && !samePath(path, opts.Root) {
			return filepath.SkipDir
		}
		scanned++
		if scanned > maxFindScanned {
			truncated = true
			return filepath.SkipAll
		}
		if scoreTokens(qTokens, tokenize(name+" "+path)) > 0 {
			if info, statErr := d.Info(); statErr == nil {
				matches = append(matches, makeEntry(path, opts.Root, info))
			}
		}
		return nil
	})
	if err != nil && !errors.Is(err, filepath.SkipAll) {
		return nil, nil, truncated, err
	}
	sort.SliceStable(matches, func(i, j int) bool {
		si := scoreTokens(qTokens, tokenize(matches[i].Name+" "+matches[i].Path))
		sj := scoreTokens(qTokens, tokenize(matches[j].Name+" "+matches[j].Path))
		if si == sj {
			return matches[i].Path < matches[j].Path
		}
		return si > sj
	})
	if len(matches) > maxFindMatches {
		matches = matches[:maxFindMatches]
		truncated = true
	}
	folders, _ := collectFolderRecords(opts)
	programsData, _ := loadProgramsFromPath(opts.ProgramsJSON)
	data := map[string]any{
		"query":                  query,
		"scanned":                scanned,
		"matches":                matches,
		"folderCandidates":       rankFolders(folders, query, opts.Root, 30),
		"existingProgramMatches": rankPrograms(programsData, query, opts.LauncherDir, 30),
	}
	return data, nil, truncated, nil
}

func commandValidate(opts options) (any, []string, bool, error) {
	var warnings []string
	pd, err := loadProgramsFromPath(opts.ProgramsJSON)
	if err != nil {
		return nil, warnings, false, err
	}
	dd, derr := loadDriversFromPath(opts.DriversJSON)
	if derr != nil {
		warnings = append(warnings, derr.Error())
	}

	programIssues, programWarnings := validatePrograms(opts, pd)
	warnings = append(warnings, programWarnings...)
	driverIssues := validateDrivers(opts, dd)
	data := map[string]any{
		"programsPath":          opts.ProgramsJSON,
		"driversPath":           opts.DriversJSON,
		"declaredTotalPrograms": pd.TotalPrograms,
		"actualTotalPrograms":   len(pd.Programs),
		"programIssueCount":     len(programIssues),
		"driverIssueCount":      len(driverIssues),
		"programIssues":         limitIssues(programIssues, 250),
		"driverIssues":          limitIssues(driverIssues, 250),
		"catalogSummary":        catalogSummary(pd),
	}
	truncated := len(programIssues) > 250 || len(driverIssues) > 250
	return data, warnings, truncated, nil
}

func commandFolders(opts options) (any, []string, bool, error) {
	folders, warnings := collectFolderRecords(opts)
	candidates := make([]folderCandidate, 0, len(folders))
	for _, f := range folders {
		candidates = append(candidates, folderCandidate{
			Path:     f.Path,
			RelPath:  relTo(opts.Root, f.Path),
			Category: f.Category,
			Source:   f.Source,
			Score:    0,
			Evidence: []string{"existing catalogue folder"},
		})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Category == candidates[j].Category {
			return candidates[i].Path < candidates[j].Path
		}
		return candidates[i].Category < candidates[j].Category
	})
	truncated := false
	if len(candidates) > maxCatalogFolders {
		candidates = candidates[:maxCatalogFolders]
		truncated = true
	}
	top, topWarnings, topTruncated := listEntries(opts.Root, opts.Root)
	warnings = append(warnings, topWarnings...)
	data := map[string]any{
		"launcherDir":    opts.LauncherDir,
		"programsPath":   opts.ProgramsJSON,
		"driversPath":    opts.DriversJSON,
		"topLevel":       top,
		"catalogFolders": candidates,
		"knownCenters":   knownCenters(opts.Root),
	}
	return data, warnings, truncated || topTruncated, nil
}

func loadProgramsFromPath(path string) (*programs.ProgramsData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read programs.json: %w", err)
	}
	var pd programs.ProgramsData
	if err := json.Unmarshal(data, &pd); err != nil {
		return nil, fmt.Errorf("cannot parse programs.json: %w", err)
	}
	return &pd, nil
}

func loadDriversFromPath(path string) (*drivers.DriversData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read drivers.json: %w", err)
	}
	var dd drivers.DriversData
	if err := json.Unmarshal(data, &dd); err != nil {
		return nil, fmt.Errorf("cannot parse drivers.json: %w", err)
	}
	return &dd, nil
}

func collectFolderRecords(opts options) ([]folderRecord, []string) {
	var warnings []string
	seen := map[string]bool{}
	var out []folderRecord
	add := func(path, category, source string) {
		path = filepath.Clean(path)
		if path == "." || path == "" || seen[strings.ToLower(path)] {
			return
		}
		seen[strings.ToLower(path)] = true
		out = append(out, folderRecord{Path: path, Category: category, Source: source, Tokens: tokenize(path + " " + category)})
	}
	if pd, err := loadProgramsFromPath(opts.ProgramsJSON); err == nil {
		for _, p := range pd.Programs {
			if strings.TrimSpace(p.FolderPath) != "" {
				add(resolveCatalogPath(opts.LauncherDir, p.FolderPath), p.Category, "programs.json")
			}
			for _, v := range p.Variants {
				if strings.TrimSpace(v.FolderPath) != "" {
					add(resolveCatalogPath(opts.LauncherDir, v.FolderPath), p.Category, "programs.json variant")
				}
			}
		}
	} else {
		warnings = append(warnings, err.Error())
	}
	if dd, err := loadDriversFromPath(opts.DriversJSON); err == nil {
		for typeKey, dt := range dd.Types {
			for brandKey, brand := range dt.Brands {
				for _, series := range brand.Series {
					if strings.TrimSpace(series.FolderPath) != "" {
						cat := strings.TrimSpace(dt.Title + " / " + brand.Title)
						if brand.Title == "" {
							cat = typeKey + " / " + brandKey
						}
						add(resolveCatalogPath(opts.LauncherDir, series.FolderPath), cat, "drivers.json")
					}
				}
			}
		}
	} else {
		warnings = append(warnings, err.Error())
	}
	for _, c := range knownCenters(opts.Root) {
		add(c.Path, c.Label, "known center")
	}
	return out, warnings
}

type knownCenter struct {
	Label  string `json:"label"`
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
}

func knownCenters(root string) []knownCenter {
	names := []string{"_Launcher", "_Inbox", "Centrum Narzędzi i Oprogramowania", "Centrum Gier i Modyfikacji"}
	centers := make([]knownCenter, 0, len(names))
	for _, name := range names {
		p := filepath.Join(root, name)
		centers = append(centers, knownCenter{Label: name, Path: p, Exists: fileExists(p)})
	}
	return centers
}

func validatePrograms(opts options, pd *programs.ProgramsData) ([]validationIssue, []string) {
	var issues []validationIssue
	var warnings []string
	if pd.TotalPrograms != 0 && pd.TotalPrograms != len(pd.Programs) {
		warnings = append(warnings, fmt.Sprintf("totalPrograms mismatch: declared %d, actual %d", pd.TotalPrograms, len(pd.Programs)))
	}
	seen := map[string]bool{}
	for _, p := range pd.Programs {
		key := strings.ToLower(strings.TrimSpace(p.Name))
		if key == "" {
			issues = append(issues, validationIssue{Name: p.Name, Category: p.Category, Field: "name", Message: "empty program name"})
		} else if seen[key] {
			issues = append(issues, validationIssue{Name: p.Name, Category: p.Category, Field: "name", Message: "duplicate program name"})
		}
		seen[key] = true

		if p.Type == "link" {
			if p.ExternalURL != "" {
				continue
			}
			if p.LocalPath == "" {
				issues = append(issues, validationIssue{Name: p.Name, Category: p.Category, Field: "localPath", Message: "link has no localPath or externalUrl"})
				continue
			}
			if strings.HasPrefix(strings.ReplaceAll(p.LocalPath, "\\", "/"), "user_pages/") {
				path := filepath.Join(opts.LauncherDir, "data", filepath.FromSlash(p.LocalPath))
				if !fileExists(path) {
					issues = append(issues, validationIssue{Name: p.Name, Category: p.Category, Field: "localPath", Path: path, Message: "user page file not found"})
				}
			} else {
				warnings = append(warnings, "local link may be embedded or legacy; not checked on disk: "+p.Name+" -> "+p.LocalPath)
			}
			continue
		}

		if strings.TrimSpace(p.FolderPath) != "" {
			path := resolveCatalogPath(opts.LauncherDir, p.FolderPath)
			if !fileExists(path) {
				issues = append(issues, validationIssue{Name: p.Name, Category: p.Category, Field: "folderPath", Path: path, Message: "folderPath not found"})
			}
			if p.InstallType == "installer" {
				if p.InstallerFile != "" {
					ip := filepath.Join(path, p.InstallerFile)
					if !fileExists(ip) {
						issues = append(issues, validationIssue{Name: p.Name, Category: p.Category, Field: "installerFile", Path: ip, Message: "installerFile not found"})
					}
				} else if fileExists(path) && len(findInstallerCandidates(path, 1)) == 0 {
					issues = append(issues, validationIssue{Name: p.Name, Category: p.Category, Field: "installType", Path: path, Message: "installType=installer but no installer candidate found"})
				}
			}
			if p.LocalPath != "" {
				lp := filepath.Join(path, p.LocalPath)
				if !fileExists(lp) {
					issues = append(issues, validationIssue{Name: p.Name, Category: p.Category, Field: "localPath", Path: lp, Message: "localPath not found inside folderPath"})
				}
			}
		}
		for _, v := range p.Variants {
			if strings.TrimSpace(v.FolderPath) == "" {
				continue
			}
			vPath := resolveCatalogPath(opts.LauncherDir, v.FolderPath)
			if !fileExists(vPath) {
				issues = append(issues, validationIssue{Name: p.Name + " / " + v.Name, Category: p.Category, Field: "variants.folderPath", Path: vPath, Message: "variant folderPath not found"})
			}
			if v.InstallerFile != "" {
				ip := filepath.Join(vPath, v.InstallerFile)
				if !fileExists(ip) {
					issues = append(issues, validationIssue{Name: p.Name + " / " + v.Name, Category: p.Category, Field: "variants.installerFile", Path: ip, Message: "variant installerFile not found"})
				}
			}
		}
	}
	return issues, warnings
}

func validateDrivers(opts options, dd *drivers.DriversData) []validationIssue {
	if dd == nil {
		return nil
	}
	var issues []validationIssue
	for typeKey, dt := range dd.Types {
		for brandKey, brand := range dt.Brands {
			for _, series := range brand.Series {
				path := resolveCatalogPath(opts.LauncherDir, series.FolderPath)
				if !fileExists(path) {
					issues = append(issues, validationIssue{Name: series.Name, Category: typeKey + "/" + brandKey, Field: "folderPath", Path: path, Message: "driver folderPath not found"})
				}
			}
		}
	}
	return issues
}

func catalogSummary(pd *programs.ProgramsData) map[string]any {
	categories := map[string]int{}
	installTypes := map[string]int{}
	profiles := map[string]int{}
	for _, p := range pd.Programs {
		categories[p.Category]++
		installTypes[p.InstallType]++
		for _, profile := range p.ProfileRelevant {
			profiles[profile]++
		}
	}
	return map[string]any{"categories": categories, "installTypes": installTypes, "profiles": profiles}
}

func listEntries(path, root string) ([]fileEntry, []string, bool) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		if err == nil {
			return []fileEntry{makeEntry(path, root, info)}, nil, false
		}
		return nil, []string{err.Error()}, false
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, []string{err.Error()}, false
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDir() == entries[j].IsDir() {
			return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
		}
		return entries[i].IsDir()
	})
	truncated := false
	if len(entries) > maxContextEntries {
		entries = entries[:maxContextEntries]
		truncated = true
	}
	result := make([]fileEntry, 0, len(entries))
	for _, e := range entries {
		if info, statErr := e.Info(); statErr == nil {
			result = append(result, makeEntry(filepath.Join(path, e.Name()), root, info))
		}
	}
	return result, nil, truncated
}

func scanStats(path string) (map[string]any, []string, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return map[string]any{}, nil, false
	}
	stats := map[string]any{"files": 0, "dirs": 0, "bytes": int64(0), "extensions": map[string]int{}}
	exts := map[string]int{}
	var evidence []string
	truncated := false
	seen := 0
	addEvidence := func(item string) {
		if len(evidence) >= 80 {
			return
		}
		for _, existing := range evidence {
			if existing == item {
				return
			}
		}
		evidence = append(evidence, item)
	}
	if !info.IsDir() {
		stats["files"] = 1
		stats["bytes"] = info.Size()
		stats["extensions"] = map[string]int{strings.ToLower(filepath.Ext(path)): 1}
		addEvidence(filepath.Base(path))
		return stats, evidence, false
	}
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && shouldSkipDir(d.Name()) && p != path {
			return filepath.SkipDir
		}
		seen++
		if seen > maxContextWalk {
			truncated = true
			return filepath.SkipAll
		}
		if d.IsDir() {
			stats["dirs"] = stats["dirs"].(int) + 1
			return nil
		}
		if info, statErr := d.Info(); statErr == nil {
			stats["files"] = stats["files"].(int) + 1
			stats["bytes"] = stats["bytes"].(int64) + info.Size()
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext != "" {
			exts[ext]++
		}
		lower := strings.ToLower(d.Name())
		if isEvidenceName(lower) {
			addEvidence(p)
		}
		return nil
	})
	stats["extensions"] = exts
	stats["scanned"] = seen
	return stats, evidence, truncated
}

func rankFolders(records []folderRecord, query, root string, limit int) []folderCandidate {
	qTokens := tokenize(query)
	if len(qTokens) == 0 {
		return nil
	}
	var out []folderCandidate
	for _, rec := range records {
		score := scoreTokens(qTokens, rec.Tokens)
		if score == 0 {
			continue
		}
		evidence := []string{"matched tokens: " + strings.Join(intersectTokens(qTokens, rec.Tokens), ", ")}
		if fileExists(rec.Path) {
			evidence = append(evidence, "folder exists")
		} else {
			evidence = append(evidence, "catalog path currently missing")
		}
		out = append(out, folderCandidate{Path: rec.Path, RelPath: relTo(root, rec.Path), Category: rec.Category, Source: rec.Source, Score: score, Evidence: evidence})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Path < out[j].Path
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func rankPrograms(pd *programs.ProgramsData, query, launcherDir string, limit int) []programMatch {
	if pd == nil {
		return nil
	}
	qTokens := tokenize(query)
	if len(qTokens) == 0 {
		return nil
	}
	var out []programMatch
	for _, p := range pd.Programs {
		text := p.Name + " " + p.Description + " " + p.Category + " " + strings.Join(p.Tags, " ") + " " + p.FolderPath + " " + p.LocalPath
		score := scoreTokens(qTokens, tokenize(text))
		if score == 0 {
			continue
		}
		path := ""
		if p.FolderPath != "" {
			path = resolveCatalogPath(launcherDir, p.FolderPath)
		}
		out = append(out, programMatch{Name: p.Name, Category: p.Category, InstallType: p.InstallType, FolderPath: p.FolderPath, Path: path, Score: score, Evidence: []string{"matched tokens: " + strings.Join(intersectTokens(qTokens, tokenize(text)), ", ")}})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Name < out[j].Name
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func suggestedQuestions(evidence []string, folders []folderCandidate, matches []programMatch) []string {
	var qs []string
	if len(folders) > 1 && folders[0].Score == folders[1].Score {
		qs = append(qs, "Który z podobnych folderów katalogu wybrać?")
	}
	if len(matches) > 0 {
		qs = append(qs, "Czy to aktualizacja istniejącego wpisu: "+matches[0].Name+"?")
	}
	archiveCount := 0
	installerCount := 0
	for _, e := range evidence {
		lower := strings.ToLower(e)
		if isArchiveExt(filepath.Ext(lower)) {
			archiveCount++
		}
		if isInstallerExt(filepath.Ext(lower)) {
			installerCount++
		}
	}
	if archiveCount > 1 || (archiveCount > 0 && installerCount > 0) {
		qs = append(qs, "Czy paczka ma zostać rozpakowana i posegregowana osobno, czy zostać jako collection/other?")
	}
	return qs
}

func findInstallerCandidates(path string, limit int) []string {
	var out []string
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		lower := strings.ToLower(d.Name())
		if strings.Contains(lower, "uninstall") || strings.Contains(lower, "unins") {
			return nil
		}
		if isInstallerExt(filepath.Ext(lower)) {
			out = append(out, p)
		}
		if len(out) >= limit {
			return filepath.SkipAll
		}
		return nil
	})
	return out
}

func makeEntry(path, root string, info os.FileInfo) fileEntry {
	return fileEntry{Name: info.Name(), Path: filepath.Clean(path), RelPath: relTo(root, path), Kind: kindOf(info), Ext: strings.ToLower(filepath.Ext(info.Name())), Size: info.Size(), Modified: info.ModTime()}
}

func kindOf(info os.FileInfo) string {
	if info.IsDir() {
		return "dir"
	}
	return "file"
}

func resolveInputPath(root, input string) string {
	input = strings.Trim(strings.TrimSpace(input), "\"")
	input = filepath.FromSlash(input)
	if filepath.IsAbs(input) {
		return filepath.Clean(input)
	}
	return filepath.Clean(filepath.Join(root, input))
}

func resolveCatalogPath(launcherDir, rawPath string) string {
	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		return ""
	}
	pathPart := filepath.FromSlash(rawPath)
	if filepath.IsAbs(pathPart) {
		return filepath.Clean(pathPart)
	}

	stripped := pathPart
	for strings.HasPrefix(stripped, ".."+string(os.PathSeparator)) || strings.HasPrefix(stripped, "..\\") || strings.HasPrefix(stripped, "../") {
		stripped = strings.TrimPrefix(stripped, ".."+string(os.PathSeparator))
		stripped = strings.TrimPrefix(stripped, "..\\")
		stripped = strings.TrimPrefix(stripped, "../")
	}

	drive := system.DriveOf(launcherDir)
	candidates := []string{
		filepath.Join(drive+"\\", stripped),
		filepath.Join(launcherDir, pathPart),
		filepath.Join(launcherDir, "..", pathPart),
		filepath.Join(launcherDir, "..", "..", stripped),
	}
	for _, c := range candidates {
		c = filepath.Clean(c)
		if fileExists(c) {
			return c
		}
	}
	return filepath.Clean(candidates[0])
}

func relTo(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func samePath(a, b string) bool {
	aa, _ := filepath.Abs(a)
	bb, _ := filepath.Abs(b)
	return strings.EqualFold(filepath.Clean(aa), filepath.Clean(bb))
}

func shouldSkipDir(name string) bool {
	lower := strings.ToLower(name)
	switch lower {
	case "$recycle.bin", "system volume information", ".git", "node_modules", ".webview_data", "logs":
		return true
	}
	return false
}

func isEvidenceName(lower string) bool {
	ext := filepath.Ext(lower)
	if isInstallerExt(ext) || isArchiveExt(ext) {
		return true
	}
	return strings.Contains(lower, "readme") || strings.Contains(lower, "install") || strings.Contains(lower, "setup") || strings.Contains(lower, "opis") || strings.Contains(lower, "license")
}

func isInstallerExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".exe", ".msi", ".msixbundle", ".appx", ".appxbundle", ".bat", ".cmd":
		return true
	}
	return false
}

func isArchiveExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".zip", ".7z", ".rar", ".tar", ".gz", ".xz", ".bz2":
		return true
	}
	return false
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	replacer := strings.NewReplacer("ą", "a", "ć", "c", "ę", "e", "ł", "l", "ń", "n", "ó", "o", "ś", "s", "ż", "z", "ź", "z")
	s = replacer.Replace(s)
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	seen := map[string]bool{}
	var out []string
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if len(f) < 2 || seen[f] {
			continue
		}
		seen[f] = true
		out = append(out, f)
	}
	return out
}

func scoreTokens(query, haystack []string) int {
	if len(query) == 0 || len(haystack) == 0 {
		return 0
	}
	score := 0
	for _, q := range query {
		for _, h := range haystack {
			if q == h {
				score += 10
				break
			}
			if len(q) >= 4 && strings.Contains(h, q) {
				score += 4
				break
			}
			if len(h) >= 4 && strings.Contains(q, h) {
				score += 2
				break
			}
		}
	}
	return score
}

func intersectTokens(a, b []string) []string {
	set := map[string]bool{}
	for _, x := range b {
		set[x] = true
	}
	var out []string
	for _, x := range a {
		if set[x] {
			out = append(out, x)
		}
	}
	return out
}

func entryNames(entries []fileEntry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name)
	}
	return out
}

func limitIssues(in []validationIssue, limit int) []validationIssue {
	if len(in) <= limit {
		return in
	}
	return in[:limit]
}

func usageData() map[string]any {
	return map[string]any{
		"usage": "disk-helper.exe [--root <drive>:\\] [--launcher <drive>:\\_Launcher] <command> [args]",
		"commands": map[string]string{
			"context <path>": "Read facts about a file/folder and return entries, stats, evidence, candidates.",
			"find <query>":   "Search file/folder names and catalogue entries.",
			"validate":       "Validate runtime programs.json and drivers.json paths.",
			"folders":        "List known top-level and catalogue folders.",
		},
		"readOnly": true,
	}
}

func envelopeFor(cmd string, args []string, opts options, warnings []string, errs []string, ok bool, truncated bool, data any) envelope {
	return envelope{Tool: toolName, SchemaVersion: schemaVersion, Command: commandMeta{Name: cmd, Args: args}, Root: opts.Root, OK: ok, Warnings: dedupeStrings(warnings), Errors: dedupeStrings(errs), Truncated: truncated, Data: data}
}

func dedupeStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func writeEnvelope(env envelope) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(env)
}
