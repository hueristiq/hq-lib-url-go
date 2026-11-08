package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_removeDotSegments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		cleaned string
	}{
		{"/a/b/c/./../../g", "/a/b/c/./../../g", "/a/g"},
		{"mid/content=5/../6", "mid/content=5/../6", "mid/6"},
		{"../g", "../g", "g"},
		{"./g", "./g", "g"},
		{"/./g", "/./g", "/g"},
		{"/.", "/.", "/"},
		{"/..", "/..", "/"},
		{".", ".", ""},
		{"..", "..", ""},
		{"g.", "g.", "g."},
		{"empty path", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.cleaned, removeDotSegments(tt.path))
		})
	}
}
