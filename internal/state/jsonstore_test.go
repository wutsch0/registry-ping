package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempStore(t *testing.T) *JSONStateStore {
	t.Helper()
	dir := t.TempDir()
	return NewJSONStateStore(filepath.Join(dir, "state.json"))
}

func TestJSONStateStore_NonExistentFile(t *testing.T) {
	s := tempStore(t)

	st, found, err := s.Load("php:8.2.30-fpm")
	require.NoError(t, err)
	assert.False(t, found)
	assert.Zero(t, st)
}

func TestJSONStateStore_RoundTrip(t *testing.T) {
	s := tempStore(t)
	key := "php:8.2.30-fpm"
	ts := time.Date(2026, 2, 4, 17, 56, 28, 0, time.UTC)

	err := s.Save(key, ImageState{LastPushed: ts})
	require.NoError(t, err)

	st, found, err := s.Load(key)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, ts, st.LastPushed)
}

func TestJSONStateStore_MultipleKeys(t *testing.T) {
	s := tempStore(t)

	ts1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	require.NoError(t, s.Save("php:8.2.30-fpm", ImageState{LastPushed: ts1}))
	require.NoError(t, s.Save("nginx:1.25-alpine", ImageState{LastPushed: ts2}))

	st1, found1, err := s.Load("php:8.2.30-fpm")
	require.NoError(t, err)
	assert.True(t, found1)
	assert.Equal(t, ts1, st1.LastPushed)

	st2, found2, err := s.Load("nginx:1.25-alpine")
	require.NoError(t, err)
	assert.True(t, found2)
	assert.Equal(t, ts2, st2.LastPushed)
}

func TestJSONStateStore_AtomicRename(t *testing.T) {
	s := tempStore(t)
	key := "php:8.2.30-fpm"
	ts := time.Date(2026, 2, 4, 17, 56, 28, 0, time.UTC)

	require.NoError(t, s.Save(key, ImageState{LastPushed: ts}))

	// The .tmp file should not be left behind
	_, err := os.Stat(s.path + ".tmp")
	assert.True(t, os.IsNotExist(err), "tmp file should have been renamed away")
}
