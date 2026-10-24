package extractor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseHexWithin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    string
		limit uint64
		want  uint64
		ok    bool
	}{
		{"within limit", "41", maxUnicodeCodePoint, 0x41, true},
		{"at limit", "10FFFF", maxUnicodeCodePoint, maxUnicodeCodePoint, true},
		{"above limit", "110000", maxUnicodeCodePoint, 0, false},
		{"overflowing uint64", "10000000000000000", maxUnicodeCodePoint, 0, false},
		{"digit pushes over limit", "100F", 0x1000, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseHexWithin(tt.in, tt.limit)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.ok, ok)
		})
	}
}
