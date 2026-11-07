package unicode_test

import (
	"regexp"
	"testing"

	hqgourlunicode "github.com/hueristiq/hq-lib-url-go/internal/unicode"
)

func matchesClass(t *testing.T, class string, r rune) bool {
	t.Helper()

	expression, err := regexp.Compile("[" + class + "]")
	if err != nil {
		t.Fatal(err)
	}

	return expression.MatchString(string(r))
}

func TestGeneratedRangesExcludePunctuation(t *testing.T) {
	t.Parallel()

	if !matchesClass(t, hqgourlunicode.AllowedUcsCharMinusPunc, 'α') {
		t.Fatal("punctuation-free range omits Greek alpha")
	}

	if matchesClass(t, hqgourlunicode.AllowedUcsCharMinusPunc, ',') {
		t.Fatal("punctuation-free range includes a comma")
	}
}
