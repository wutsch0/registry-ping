package dockerhub

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wutscho/registry-ping/internal/registry"
)

func newTestScraper(server *httptest.Server) *DockerHubScraper {
	return NewDockerHubScraper(server.Client(), WithBaseURL(server.URL))
}

func TestFetch_OfficialImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/repositories/library/php/tags/8.2.30-fpm", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_last_pushed":"2026-02-04T17:56:28.838962Z"}`))
	}))
	defer server.Close()

	scraper := newTestScraper(server)
	ref := registry.ImageRef{Namespace: "library", Name: "php", Tag: "8.2.30-fpm"}

	info, err := scraper.Fetch(context.Background(), ref)
	require.NoError(t, err)
	assert.Equal(t, ref, info.Ref)

	want := time.Date(2026, 2, 4, 17, 56, 28, 838962000, time.UTC)
	assert.Equal(t, want, info.LastPushed)
}

func TestFetch_UserImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/repositories/myorg/myimage/tags/1.0.0", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_last_pushed":"2026-01-15T10:00:00Z"}`))
	}))
	defer server.Close()

	scraper := newTestScraper(server)
	ref := registry.ImageRef{Namespace: "myorg", Name: "myimage", Tag: "1.0.0"}

	info, err := scraper.Fetch(context.Background(), ref)
	require.NoError(t, err)
	assert.Equal(t, ref, info.Ref)
	assert.Equal(t, time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC), info.LastPushed)
}

func TestFetch_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	scraper := newTestScraper(server)
	ref := registry.ImageRef{Namespace: "library", Name: "php", Tag: "99.99.99"}

	_, err := scraper.Fetch(context.Background(), ref)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound), "expected ErrNotFound, got: %v", err)
}

func TestFetch_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := newTestScraper(server)
	ref := registry.ImageRef{Namespace: "library", Name: "php", Tag: "8.2.30-fpm"}

	_, err := scraper.Fetch(context.Background(), ref)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestCanHandle(t *testing.T) {
	scraper := NewDockerHubScraper(&http.Client{})

	tests := []struct {
		host string
		want bool
	}{
		{"", true},
		{"docker.io", true},
		{"hub.docker.com", true},
		{"ghcr.io", false},
		{"registry.example.com", false},
	}

	for _, tc := range tests {
		t.Run(tc.host, func(t *testing.T) {
			assert.Equal(t, tc.want, scraper.CanHandle(tc.host))
		})
	}
}
