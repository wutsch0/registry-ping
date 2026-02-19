package dockerhub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/wutscho/registry-ping/internal/registry"
)

// ErrNotFound is returned when the requested image tag does not exist.
var ErrNotFound = errors.New("image tag not found")

const defaultBaseURL = "https://hub.docker.com"

// DockerHubScraper fetches image metadata from the Docker Hub public REST API.
type DockerHubScraper struct {
	client  *http.Client
	baseURL string
}

// Option is a functional option for DockerHubScraper.
type Option func(*DockerHubScraper)

// WithBaseURL overrides the Docker Hub base URL (useful for testing with httptest).
func WithBaseURL(url string) Option {
	return func(s *DockerHubScraper) {
		s.baseURL = url
	}
}

// NewDockerHubScraper creates a new DockerHubScraper using the given HTTP client.
func NewDockerHubScraper(client *http.Client, opts ...Option) *DockerHubScraper {
	s := &DockerHubScraper{
		client:  client,
		baseURL: defaultBaseURL,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// CanHandle reports whether this scraper handles the given host.
func (s *DockerHubScraper) CanHandle(host string) bool {
	switch host {
	case "", "docker.io", "hub.docker.com":
		return true
	}
	return false
}

type tagResponse struct {
	TagLastPushed time.Time `json:"tag_last_pushed"`
}

// Fetch retrieves the tag_last_pushed timestamp for the image from Docker Hub.
func (s *DockerHubScraper) Fetch(ctx context.Context, ref registry.ImageRef) (registry.ImageInfo, error) {
	url := fmt.Sprintf("%s/v2/repositories/%s/%s/tags/%s",
		s.baseURL, ref.Namespace, ref.Name, ref.Tag)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return registry.ImageInfo{}, fmt.Errorf("dockerhub: create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return registry.ImageInfo{}, fmt.Errorf("dockerhub: fetch %s: %w", ref, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return registry.ImageInfo{}, fmt.Errorf("dockerhub: %s: %w", ref, ErrNotFound)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return registry.ImageInfo{}, fmt.Errorf("dockerhub: %s: unexpected status %d", ref, resp.StatusCode)
	}

	var data tagResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return registry.ImageInfo{}, fmt.Errorf("dockerhub: decode response for %s: %w", ref, err)
	}

	return registry.ImageInfo{
		Ref:        ref,
		LastPushed: data.TagLastPushed.UTC(),
	}, nil
}
