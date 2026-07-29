package api

// Custom install profiles stored in the profiles store.

import (
	"encoding/json"

	"launcher/internal/programs"
)

func (a *API) GetCustomProfiles() string {
	if a.ProfilesStore == nil {
		return `[]`
	}
	data, _ := json.Marshal(a.ProfilesStore.GetAll())
	return string(data)
}

func (a *API) SaveCustomProfile(profileJSON string) string {
	if a.ProfilesStore == nil {
		return `{"error": "Profiles store not initialized"}`
	}

	var p programs.CustomProfile
	if err := json.Unmarshal([]byte(profileJSON), &p); err != nil {
		return `{"error": "Invalid profile JSON"}`
	}

	if err := a.ProfilesStore.Save(p); err != nil {
		return jsonError(err.Error())
	}
	return `{"success": true}`
}

func (a *API) DeleteCustomProfile(id string) string {
	if a.ProfilesStore == nil {
		return `{"error": "Profiles store not initialized"}`
	}

	if err := a.ProfilesStore.Delete(id); err != nil {
		return jsonError(err.Error())
	}
	return `{"success": true}`
}

func (a *API) AddProgramToProfile(profileID, programName string) string {
	if a.ProfilesStore == nil {
		return `{"error": "Profiles store not initialized"}`
	}

	if err := a.ProfilesStore.AddProgramToProfile(profileID, programName); err != nil {
		return jsonError(err.Error())
	}
	return `{"success": true}`
}

func (a *API) RemoveProgramFromProfile(profileID, programName string) string {
	if a.ProfilesStore == nil {
		return `{"error": "Profiles store not initialized"}`
	}

	if err := a.ProfilesStore.RemoveProgramFromProfile(profileID, programName); err != nil {
		return jsonError(err.Error())
	}
	return `{"success": true}`
}
