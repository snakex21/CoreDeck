package windowstate

import (
	"encoding/json"
	"os"

	"launcher/internal/safeio"
)

const (
	MinWidth  = 800
	MinHeight = 600
	MaxWidth  = 7680
	MaxHeight = 4320
)

type State struct {
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Maximized bool `json:"maximized"`
}

func Load(path string, fallback State) State {
	data, err := os.ReadFile(path)
	if err != nil {
		return normalize(fallback, fallback)
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return normalize(fallback, fallback)
	}
	return normalize(state, fallback)
}

func Save(path string, state State) error {
	state = normalize(state, State{Width: 1400, Height: 900, Maximized: true})
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return safeio.WriteFileAtomic(path, data, 0o644)
}

func normalize(state, fallback State) State {
	if fallback.Width < MinWidth || fallback.Width > MaxWidth {
		fallback.Width = 1400
	}
	if fallback.Height < MinHeight || fallback.Height > MaxHeight {
		fallback.Height = 900
	}
	if state.Width < MinWidth || state.Width > MaxWidth {
		state.Width = fallback.Width
	}
	if state.Height < MinHeight || state.Height > MaxHeight {
		state.Height = fallback.Height
	}
	return state
}
