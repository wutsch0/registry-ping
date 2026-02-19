package registry

import (
	"fmt"
	"strings"
	"time"
)

// ImageRef identifies a specific tagged image in a container registry.
type ImageRef struct {
	Host      string // "" = Docker Hub
	Namespace string // "library" for official Docker Hub images
	Name      string
	Tag       string
}

// ParseImageRef parses a string like "php:8.2.30-fpm", "myorg/img:1.0", or
// "ghcr.io/org/img:latest" into an ImageRef. A tag is required.
func ParseImageRef(s string) (ImageRef, error) {
	// Split off tag at last ':'
	lastColon := strings.LastIndex(s, ":")
	if lastColon < 0 {
		return ImageRef{}, fmt.Errorf("image ref %q: tag required (use name:tag)", s)
	}
	path := s[:lastColon]
	tag := s[lastColon+1:]
	if tag == "" {
		return ImageRef{}, fmt.Errorf("image ref %q: tag must not be empty", s)
	}

	// Split path into segments
	segments := strings.Split(path, "/")

	var host, namespace, name string

	// First segment is a host if it contains a '.' or is "localhost"
	if len(segments) > 0 && (strings.Contains(segments[0], ".") || segments[0] == "localhost") {
		host = segments[0]
		segments = segments[1:]
	}

	switch len(segments) {
	case 1:
		if host == "" {
			namespace = "library"
		}
		name = segments[0]
	case 2:
		namespace = segments[0]
		name = segments[1]
	default:
		return ImageRef{}, fmt.Errorf("image ref %q: unsupported path format", s)
	}

	if name == "" {
		return ImageRef{}, fmt.Errorf("image ref %q: name must not be empty", s)
	}

	return ImageRef{
		Host:      host,
		Namespace: namespace,
		Name:      name,
		Tag:       tag,
	}, nil
}

// String returns a human-readable image reference. "library/" is omitted for
// Docker Hub official images. Used as the state file key.
func (r ImageRef) String() string {
	var b strings.Builder
	if r.Host != "" {
		b.WriteString(r.Host)
		b.WriteByte('/')
	}
	if r.Namespace != "" && r.Namespace != "library" {
		b.WriteString(r.Namespace)
		b.WriteByte('/')
	}
	b.WriteString(r.Name)
	b.WriteByte(':')
	b.WriteString(r.Tag)
	return b.String()
}

// ImageInfo holds the fetched metadata for an image tag.
type ImageInfo struct {
	Ref        ImageRef
	LastPushed time.Time
}
