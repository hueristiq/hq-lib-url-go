package tlds

import (
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validTLD is the structural shape every generated entry must satisfy:
// no whitespace, no leading or trailing dot, no empty labels, and ASCII
// letters lowercase. Unicode (U-label) entries such as "andøy.no" are
// permitted.
var validTLD = regexp.MustCompile(`^\S+$`)

func assertValidEntry(t *testing.T, entry string) {
	t.Helper()

	assert.Regexp(t, validTLD, entry)
	assert.NotContains(t, entry, "..", "entry %q must not have empty labels", entry)
	assert.False(t, strings.HasPrefix(entry, "."), "entry %q must not start with a dot", entry)
	assert.False(t, strings.HasSuffix(entry, "."), "entry %q must not end with a dot", entry)

	for _, r := range entry {
		if r < utf8.RuneSelf && r >= 'A' && r <= 'Z' {
			assert.Failf(t, "entry %q must not contain ASCII uppercase", entry)

			return
		}
	}
}

func TestOfficialStructure(t *testing.T) {
	t.Parallel()

	require.NotEmpty(t, Official)
	assert.True(t, slices.IsSorted(Official), "Official must stay sorted")
	assert.Equal(t, slices.Compact(slices.Clone(Official)), Official, "Official must not contain duplicates")

	for _, entry := range Official {
		assertValidEntry(t, entry)
	}
}

func TestPseudoStructure(t *testing.T) {
	t.Parallel()

	require.NotEmpty(t, Pseudo)
	assert.True(t, slices.IsSorted(Pseudo), "Pseudo must stay sorted")
	assert.Equal(t, slices.Compact(slices.Clone(Pseudo)), Pseudo, "Pseudo must not contain duplicates")

	for _, entry := range Pseudo {
		assertValidEntry(t, entry)
	}
}

func TestListsAreDisjoint(t *testing.T) {
	t.Parallel()

	official := map[string]struct{}{}

	for _, entry := range Official {
		official[entry] = struct{}{}
	}

	for _, entry := range Pseudo {
		_, duplicate := official[entry]

		assert.False(t, duplicate, "Pseudo entry %q also present in Official", entry)
	}
}

func TestKnownEntries(t *testing.T) {
	t.Parallel()

	// Spot checks that anchor the lists to their data sources: an IANA TLD,
	// a public suffix, and the well-known pseudo TLDs.
	assert.Contains(t, Official, "com")
	assert.Contains(t, Official, "co.uk")
	assert.Contains(t, Pseudo, "local")
	assert.Contains(t, Pseudo, "localhost")
}
