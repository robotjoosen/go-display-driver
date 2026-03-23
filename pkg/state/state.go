package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type DisplaySnapshot struct {
	ScreenType string `json:"screen_type"`
	ListIndex  int    `json:"list_index"`
}

type State struct {
	SelectedIndex int                     `json:"selected_index"`
	Displays      map[int]DisplaySnapshot `json:"displays"`
}

func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func Save(s *State, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
