package notify

import (
	"time"

	"github.com/wutscho/registry-ping/internal/registry"
)

// ChangeEvent describes a detected change for a single image tag.
type ChangeEvent struct {
	Ref         registry.ImageRef
	OldPushed   time.Time
	NewPushed   time.Time
	IsFirstSeen bool
}

// Notifier is called for each detected change.
type Notifier interface {
	Notify(event ChangeEvent) error
}
