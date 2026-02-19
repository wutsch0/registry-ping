package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// JSONStateStore persists image states as a flat JSON object on disk.
// Writes are atomic (write to .tmp then os.Rename).
type JSONStateStore struct {
	path string
}

// NewJSONStateStore creates a JSONStateStore that reads/writes the given file path.
func NewJSONStateStore(path string) *JSONStateStore {
	return &JSONStateStore{path: path}
}

func (s *JSONStateStore) load() (map[string]ImageState, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]ImageState), nil
		}
		return nil, fmt.Errorf("state: read %s: %w", s.path, err)
	}
	var m map[string]ImageState
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("state: parse %s: %w", s.path, err)
	}
	return m, nil
}

// Load retrieves the stored state for key.
func (s *JSONStateStore) Load(key string) (ImageState, bool, error) {
	m, err := s.load()
	if err != nil {
		return ImageState{}, false, err
	}
	st, ok := m[key]
	return st, ok, nil
}

// Save writes the state for key atomically.
func (s *JSONStateStore) Save(key string, st ImageState) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	m[key] = st

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("state: marshal: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("state: write tmp %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("state: rename %s -> %s: %w", tmp, s.path, err)
	}
	return nil
}
