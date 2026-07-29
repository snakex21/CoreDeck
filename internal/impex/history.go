package impex

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	historyFileName  = ".launcher_export_history.json"
	historyMaxEntries = 50
)

// ExportRecord to pojedynczy wpis historii eksportu.
type ExportRecord struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"createdAt"` // RFC3339
	TargetPath   string `json:"targetPath"`
	Format       string `json:"format"` // "folder" albo "zip"
	ProgramCount int    `json:"programCount"`
	DriverCount  int    `json:"driverCount"`
	IncludeFiles bool   `json:"includeFiles"`
}

// historyFile zwraca ścieżkę do pliku historii obok exe.
func historyFile(exeDir string) string {
	return filepath.Join(exeDir, historyFileName)
}

// newRecordID generuje ID z timestampu + krótkiego losowego sufiksu.
func newRecordID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102-150405")
	}
	return time.Now().Format("20060102-150405") + "-" + hex.EncodeToString(b)
}

// LoadHistory wczytuje historię eksportów z pliku obok exe.
// Zwraca posortowaną listę (najnowsze pierwsze). Brak pliku = pusta lista.
func LoadHistory(exeDir string) []ExportRecord {
	data, err := os.ReadFile(historyFile(exeDir))
	if err != nil {
		return []ExportRecord{}
	}
	var records []ExportRecord
	if err := json.Unmarshal(data, &records); err != nil {
		slog.Warn("export history: nie udało się sparsować pliku", "error", err)
		return []ExportRecord{}
	}
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})
	return records
}

func saveHistory(exeDir string, records []ExportRecord) {
	if records == nil {
		records = []ExportRecord{}
	}
	out, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		slog.Warn("export history: błąd marshallingu", "error", err)
		return
	}
	if err := os.WriteFile(historyFile(exeDir), out, 0644); err != nil {
		slog.Warn("export history: nie udało się zapisać pliku", "error", err)
	}
}

// AppendHistory dopisuje wpis na początek historii i ucina do limitu.
// Zwraca zaktualizowaną listę.
func AppendHistory(exeDir string, record ExportRecord) []ExportRecord {
	if record.ID == "" {
		record.ID = newRecordID()
	}
	if record.CreatedAt == "" {
		record.CreatedAt = time.Now().Format(time.RFC3339)
	}
	records := LoadHistory(exeDir)

	// Usuń ewentualny duplikat po ID (zabezpieczenie).
	filtered := records[:0]
	for _, r := range records {
		if r.ID != record.ID {
			filtered = append(filtered, r)
		}
	}
	records = append([]ExportRecord{record}, filtered...)

	if len(records) > historyMaxEntries {
		records = records[:historyMaxEntries]
	}
	saveHistory(exeDir, records)
	return records
}

// RemoveHistoryEntry usuwa wpis o danym ID z historii (plik na dysku zostaje).
// Zwraca true, jeśli wpis istniał i został usunięty.
func RemoveHistoryEntry(exeDir, id string) bool {
	records := LoadHistory(exeDir)
	found := false
	filtered := records[:0]
	for _, r := range records {
		if r.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, r)
	}
	if !found {
		return false
	}
	saveHistory(exeDir, filtered)
	return true
}

// ClearHistory czyści całą historię.
func ClearHistory(exeDir string) {
	saveHistory(exeDir, []ExportRecord{})
}
