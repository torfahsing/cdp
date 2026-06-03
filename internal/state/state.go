package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type UIDMapping struct {
	BackendNodeID int64  `json:"backendNodeId"`
	Role          string `json:"role,omitempty"`
	Name          string `json:"name,omitempty"`
}

type State struct {
	SelectedPage string                `json:"selectedPage,omitempty"`
	UIDs         map[string]UIDMapping `json:"uids,omitempty"`
}

func stateFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cdp", "state.json")
}

func Load() (*State, error) {
	data, err := os.ReadFile(stateFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return &State{}, nil
	}
	return &s, nil
}

func (s *State) Save() error {
	path := stateFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
