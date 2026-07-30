package programs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"launcher/internal/jsonutil"
	"launcher/internal/safeio"
)

// CustomProfile represents a user-created or edited profile.
type CustomProfile struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Programs    []string `json:"programs"` // list of program names
	Tweaks      []string `json:"tweaks"`   // list of tweak IDs
}

// ProfilesStore manages custom profiles persisted on disk.
type ProfilesStore struct {
	mu       sync.Mutex
	filePath string
	Profiles []CustomProfile `json:"profiles"`
}

// NewProfilesStore creates or loads a ProfilesStore from the given directory.
func NewProfilesStore(dataDir string) (*ProfilesStore, error) {
	path := filepath.Join(dataDir, "custom_profiles.json")
	store := &ProfilesStore{filePath: path}

	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("cannot read custom_profiles.json: %w", err)
		}
		if err := jsonutil.Unmarshal(data, store); err != nil {
			return nil, fmt.Errorf("cannot parse custom_profiles.json: %w", err)
		}
	}

	return store, nil
}

// save writes the store to disk.
func (s *ProfilesStore) save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return safeio.WriteFileAtomic(s.filePath, data, 0o644)
}

// GetAll returns all custom profiles.
func (s *ProfilesStore) GetAll() []CustomProfile {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Profiles == nil {
		return []CustomProfile{}
	}
	result := make([]CustomProfile, len(s.Profiles))
	for i, profile := range s.Profiles {
		result[i] = profile
		result[i].Programs = append([]string(nil), profile.Programs...)
		result[i].Tweaks = append([]string(nil), profile.Tweaks...)
	}
	return result
}

// Save creates or updates a profile.
func (s *ProfilesStore) Save(p CustomProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p.ID == "" {
		return fmt.Errorf("profile ID is required")
	}

	for i, existing := range s.Profiles {
		if existing.ID == p.ID {
			s.Profiles[i] = p
			return s.save()
		}
	}

	s.Profiles = append(s.Profiles, p)
	return s.save()
}

// Delete removes a profile by ID.
func (s *ProfilesStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.Profiles {
		if p.ID == id {
			s.Profiles = append(s.Profiles[:i], s.Profiles[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("profile not found: %s", id)
}

// AddProgramToProfile adds a program name to a profile.
func (s *ProfilesStore) AddProgramToProfile(profileID, programName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.Profiles {
		if p.ID == profileID {
			for _, name := range p.Programs {
				if name == programName {
					return nil // already there
				}
			}
			s.Profiles[i].Programs = append(s.Profiles[i].Programs, programName)
			return s.save()
		}
	}
	return fmt.Errorf("profile not found: %s", profileID)
}

// RemoveProgramFromProfile removes a program name from a profile.
func (s *ProfilesStore) RemoveProgramFromProfile(profileID, programName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.Profiles {
		if p.ID == profileID {
			for j, name := range p.Programs {
				if name == programName {
					s.Profiles[i].Programs = append(p.Programs[:j], p.Programs[j+1:]...)
					return s.save()
				}
			}
			return nil // not found, nothing to do
		}
	}
	return fmt.Errorf("profile not found: %s", profileID)
}
