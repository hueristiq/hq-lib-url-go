package tld_test

import (
	"slices"
	"testing"

	"github.com/hueristiq/hq-lib-url-go/internal/tld"
)

func TestGeneratedOfficialTLDs(t *testing.T) {
	t.Parallel()

	if !slices.Contains(tld.Official, "com") || !slices.Contains(tld.Official, "co.uk") {
		t.Fatal("official TLDs omit com or co.uk")
	}
}
