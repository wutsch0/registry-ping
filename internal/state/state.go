package state

import "time"

// ImageState holds the persisted metadata for a single image tag.
type ImageState struct {
	LastPushed time.Time `json:"last_pushed"`
}

// StateStore persists and retrieves image states by key.
// The key is typically the string representation of an ImageRef.
type StateStore interface {
	// Load retrieves the stored state for the given key.
	// Returns (state, true, nil) if found, (zero, false, nil) if not found,
	// or (zero, false, err) on I/O error.
	Load(key string) (ImageState, bool, error)
	// Save stores the state for the given key.
	Save(key string, s ImageState) error
}
