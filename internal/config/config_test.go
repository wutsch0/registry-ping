package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestLoad_Valid(t *testing.T) {
	path := writeConfig(t, `
state_file: /var/lib/registry-ping/state.json
images:
  - ref: php:8.2.30-fpm
  - ref: nginx:1.25-alpine
`)

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "/var/lib/registry-ping/state.json", cfg.StateFile)
	require.Len(t, cfg.Images, 2)
	assert.Equal(t, "php:8.2.30-fpm", cfg.Images[0].Ref)
	assert.Equal(t, "nginx:1.25-alpine", cfg.Images[1].Ref)
}

func TestLoad_DefaultStateFile(t *testing.T) {
	path := writeConfig(t, `
images:
  - ref: php:8.2.30-fpm
`)

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "state.json", cfg.StateFile)
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	require.Error(t, err)
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeConfig(t, `{invalid yaml: [}`)

	_, err := Load(path)
	require.Error(t, err)
}
