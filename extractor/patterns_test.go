package extractor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
)

func Test_SchemePatterns_caseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		text    string
	}{
		{"scheme pattern", hqgourlextractor.SchemePattern, "MAILTO:user@example.com"},
		{"known official", hqgourlextractor.KnownOfficialSchemePattern, "HTTPS://example.com"},
		{"known unofficial", hqgourlextractor.KnownUnofficialSchemePattern, "SLACK://example.com"},
		{"known no authority", hqgourlextractor.KnownNoAuthoritySchemePattern, "MAILTO:user@example.com"},
		{"known scheme", hqgourlextractor.KnownSchemePattern, "MAILTO:user@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Regexp(t, tt.pattern, tt.text)
		})
	}
}
