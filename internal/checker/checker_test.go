package checker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wutscho/registry-ping/internal/config"
	"github.com/wutscho/registry-ping/internal/notify"
	"github.com/wutscho/registry-ping/internal/registry"
	"github.com/wutscho/registry-ping/internal/state"
)

// --- mock scraper registry ---

type mockScraperRegistry struct {
	scraper registry.Scraper
	err     error
}

func (m *mockScraperRegistry) For(_ registry.ImageRef) (registry.Scraper, error) {
	return m.scraper, m.err
}

// --- mock scraper ---

type mockScraper struct {
	info registry.ImageInfo
	err  error
}

func (m *mockScraper) Fetch(_ context.Context, ref registry.ImageRef) (registry.ImageInfo, error) {
	if m.err != nil {
		return registry.ImageInfo{}, m.err
	}
	info := m.info
	info.Ref = ref
	return info, nil
}

func (m *mockScraper) CanHandle(_ string) bool { return true }

// --- mock state store ---

type mockStateStore struct {
	data    map[string]state.ImageState
	loadErr error
	saveErr error
	saved   map[string]state.ImageState
}

func newMockStore(data map[string]state.ImageState) *mockStateStore {
	if data == nil {
		data = make(map[string]state.ImageState)
	}
	return &mockStateStore{data: data, saved: make(map[string]state.ImageState)}
}

func (m *mockStateStore) Load(key string) (state.ImageState, bool, error) {
	if m.loadErr != nil {
		return state.ImageState{}, false, m.loadErr
	}
	st, ok := m.data[key]
	return st, ok, nil
}

func (m *mockStateStore) Save(key string, s state.ImageState) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved[key] = s
	return nil
}

// --- mock notifier ---

type mockNotifier struct {
	events []notify.ChangeEvent
	err    error
}

func (m *mockNotifier) Notify(event notify.ChangeEvent) error {
	if m.err != nil {
		return m.err
	}
	m.events = append(m.events, event)
	return nil
}

// --- helpers ---

var ts1 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
var ts2 = time.Date(2026, 2, 4, 17, 56, 28, 0, time.UTC)

func images(refs ...string) []config.ImageEntry {
	entries := make([]config.ImageEntry, len(refs))
	for i, r := range refs {
		entries[i] = config.ImageEntry{Ref: r}
	}
	return entries
}

// --- tests ---

func TestChecker_FirstSeen(t *testing.T) {
	scraper := &mockScraper{info: registry.ImageInfo{LastPushed: ts2}}
	reg := &mockScraperRegistry{scraper: scraper}
	store := newMockStore(nil)
	notifier := &mockNotifier{}

	c := NewChecker(reg, store, notifier)
	err := c.Run(context.Background(), images("php:8.2.30-fpm"))

	require.NoError(t, err)
	require.Len(t, notifier.events, 1)
	assert.True(t, notifier.events[0].IsFirstSeen)
	assert.Equal(t, ts2, notifier.events[0].NewPushed)
	assert.Equal(t, ts2, store.saved["php:8.2.30-fpm"].LastPushed)
}

func TestChecker_NoChange(t *testing.T) {
	scraper := &mockScraper{info: registry.ImageInfo{LastPushed: ts2}}
	reg := &mockScraperRegistry{scraper: scraper}
	store := newMockStore(map[string]state.ImageState{
		"php:8.2.30-fpm": {LastPushed: ts2},
	})
	notifier := &mockNotifier{}

	c := NewChecker(reg, store, notifier)
	err := c.Run(context.Background(), images("php:8.2.30-fpm"))

	require.NoError(t, err)
	assert.Empty(t, notifier.events, "no notification expected on no change")
	assert.Empty(t, store.saved, "no save expected on no change")
}

func TestChecker_Updated(t *testing.T) {
	scraper := &mockScraper{info: registry.ImageInfo{LastPushed: ts2}}
	reg := &mockScraperRegistry{scraper: scraper}
	store := newMockStore(map[string]state.ImageState{
		"php:8.2.30-fpm": {LastPushed: ts1},
	})
	notifier := &mockNotifier{}

	c := NewChecker(reg, store, notifier)
	err := c.Run(context.Background(), images("php:8.2.30-fpm"))

	require.NoError(t, err)
	require.Len(t, notifier.events, 1)
	assert.False(t, notifier.events[0].IsFirstSeen)
	assert.Equal(t, ts1, notifier.events[0].OldPushed)
	assert.Equal(t, ts2, notifier.events[0].NewPushed)
	assert.Equal(t, ts2, store.saved["php:8.2.30-fpm"].LastPushed)
}

func TestChecker_FetchErrorCollected(t *testing.T) {
	fetchErr := errors.New("connection refused")
	scraper := &mockScraper{err: fetchErr}
	reg := &mockScraperRegistry{scraper: scraper}
	store := newMockStore(nil)
	notifier := &mockNotifier{}

	c := NewChecker(reg, store, notifier)
	err := c.Run(context.Background(), images("php:8.2.30-fpm"))

	require.Error(t, err)
	assert.Empty(t, notifier.events)
	assert.Empty(t, store.saved)
}

func TestChecker_UnknownScraper(t *testing.T) {
	reg := &mockScraperRegistry{err: errors.New("no scraper for host")}
	store := newMockStore(nil)
	notifier := &mockNotifier{}

	c := NewChecker(reg, store, notifier)
	err := c.Run(context.Background(), images("ghcr.io/org/img:latest"))

	require.Error(t, err)
	assert.Empty(t, notifier.events)
}
