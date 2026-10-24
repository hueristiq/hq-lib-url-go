package extractor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hueristiq/hq-lib-url-go/extractor"
)

func Test_SchemePatterns_caseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		text    string
	}{
		{"scheme pattern", extractor.SchemePattern, "MAILTO:user@example.com"},
		{"known official", extractor.KnownOfficialSchemePattern, "HTTPS://example.com"},
		{"known unofficial", extractor.KnownUnofficialSchemePattern, "SLACK://example.com"},
		{"known no authority", extractor.KnownNoAuthoritySchemePattern, "MAILTO:user@example.com"},
		{"known scheme", extractor.KnownSchemePattern, "MAILTO:user@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Regexp(t, tt.pattern, tt.text)
		})
	}
}
