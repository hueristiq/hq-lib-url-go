package extractor

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	hqgourltld "github.com/hueristiq/hq-lib-url-go/internal/tld"
)

func Test_hasASCIIWordEnding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"letters", "com", true},
		{"digits and underscore", "c0_", true},
		{"uppercase", "COM", true},
		{"non-ASCII ending", "café", false},
		{"punctuation ending", "x-", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, hasASCIIWordEnding(tt.in))
		})
	}
}

func Test_splitTLDs(t *testing.T) {
	t.Parallel()

	bounded, unbounded := splitTLDs([]string{"com", "café", "", "xn--p1ai"})

	assert.Equal(t, []string{"com", "xn--p1ai"}, bounded)
	assert.Equal(t, []string{"café", ""}, unbounded)
}

func Test_boundedAndUnboundedTLDs_coverOfficialAndPseudo(t *testing.T) {
	t.Parallel()

	assert.Equal(t, len(hqgourltld.Official)+len(hqgourltld.Pseudo), len(boundedTLDs)+len(unboundedTLDs))

	for _, tld := range hqgourltld.Pseudo {
		assert.True(t, slices.Contains(boundedTLDs, tld) || slices.Contains(unboundedTLDs, tld), "%q must be classified", tld)
	}

	for _, tld := range boundedTLDs {
		assert.True(t, hasASCIIWordEnding(tld), "%q must end in an ASCII word character", tld)
	}

	for _, tld := range unboundedTLDs {
		assert.False(t, hasASCIIWordEnding(tld), "%q must not end in an ASCII word character", tld)
	}
}
