package registry

import "context"

// Scraper fetches image metadata from a container registry.
type Scraper interface {
	// Fetch retrieves the latest metadata for the given image reference.
	Fetch(ctx context.Context, ref ImageRef) (ImageInfo, error)
	// CanHandle reports whether this scraper supports the given registry host.
	// host is "" for Docker Hub.
	CanHandle(host string) bool
}
