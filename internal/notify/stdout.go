package notify

import "fmt"

// StdoutNotifier prints one line per change to stdout.
// Silent on no change (caller decides whether to call Notify).
type StdoutNotifier struct{}

// NewStdoutNotifier creates a StdoutNotifier.
func NewStdoutNotifier() *StdoutNotifier {
	return &StdoutNotifier{}
}

// Notify prints the change event to stdout.
func (n *StdoutNotifier) Notify(event ChangeEvent) error {
	if event.IsFirstSeen {
		fmt.Printf("[NEW]     %s  last_pushed=%s\n",
			event.Ref.String(),
			event.NewPushed.UTC().Format("2006-01-02T15:04:05Z"))
	} else {
		fmt.Printf("[UPDATED] %s  %s -> %s\n",
			event.Ref.String(),
			event.OldPushed.UTC().Format("2006-01-02T15:04:05Z"),
			event.NewPushed.UTC().Format("2006-01-02T15:04:05Z"))
	}
	return nil
}
