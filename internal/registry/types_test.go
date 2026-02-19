package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseImageRef(t *testing.T) {
	tests := []struct {
		input   string
		want    ImageRef
		wantErr bool
	}{
		{
			input: "php:8.2.30-fpm",
			want:  ImageRef{Host: "", Namespace: "library", Name: "php", Tag: "8.2.30-fpm"},
		},
		{
			input: "myorg/myimage:1.0.0",
			want:  ImageRef{Host: "", Namespace: "myorg", Name: "myimage", Tag: "1.0.0"},
		},
		{
			input: "ghcr.io/org/img:latest",
			want:  ImageRef{Host: "ghcr.io", Namespace: "org", Name: "img", Tag: "latest"},
		},
		{
			input: "nginx:1.25-alpine",
			want:  ImageRef{Host: "", Namespace: "library", Name: "nginx", Tag: "1.25-alpine"},
		},
		{
			input:   "php",
			wantErr: true,
		},
		{
			input:   "php:",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseImageRef(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestImageRefString(t *testing.T) {
	tests := []struct {
		ref  ImageRef
		want string
	}{
		{
			ref:  ImageRef{Host: "", Namespace: "library", Name: "php", Tag: "8.2.30-fpm"},
			want: "php:8.2.30-fpm",
		},
		{
			ref:  ImageRef{Host: "", Namespace: "myorg", Name: "myimage", Tag: "1.0.0"},
			want: "myorg/myimage:1.0.0",
		},
		{
			ref:  ImageRef{Host: "ghcr.io", Namespace: "org", Name: "img", Tag: "latest"},
			want: "ghcr.io/org/img:latest",
		},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.ref.String())
		})
	}
}
