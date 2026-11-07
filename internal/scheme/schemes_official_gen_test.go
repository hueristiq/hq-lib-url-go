package scheme_test

import (
	"slices"
	"testing"

	"github.com/hueristiq/hq-lib-url-go/internal/scheme"
)

func TestGeneratedOfficialSchemes(t *testing.T) {
	t.Parallel()

	if !slices.Contains(scheme.Official, "https") {
		t.Fatal("official schemes omit https")
	}

	if slices.Contains(scheme.Official, "not-a-real-scheme") {
		t.Fatal("official schemes contain an invalid entry")
	}
}
