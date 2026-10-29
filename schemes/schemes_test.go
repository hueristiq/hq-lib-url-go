package schemes_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	hqgourlschemes "github.com/hueristiq/hq-lib-url-go/schemes"
)

func Test_Official_excludesObsoleteSchemes(t *testing.T) {
	t.Parallel()

	obsolete := []string{
		"fax",
		"mailserver",
		"prospero",
		"shttp",
		"snews",
		"videotex",
		"wais",
		"z39.50",
		"z39.50r",
		"z39.50s",
	}

	for _, scheme := range obsolete {
		_, found := slices.BinarySearch(hqgourlschemes.Official, scheme)

		assert.False(t, found, "%q must not be listed as an official scheme", scheme)
	}
}

func Test_Official_includesActiveSchemes(t *testing.T) {
	t.Parallel()

	for _, scheme := range []string{"http", "https", "mailto", "ftp"} {
		_, found := slices.BinarySearch(hqgourlschemes.Official, scheme)

		assert.True(t, found, "%q must be listed as an official scheme", scheme)
	}
}

func Test_Official_isSortedAndDeduplicated(t *testing.T) {
	t.Parallel()

	assert.True(t, slices.IsSorted(hqgourlschemes.Official), "Official must be sorted ascending")
	assert.Len(t, slices.Compact(slices.Clone(hqgourlschemes.Official)), len(hqgourlschemes.Official), "Official must not contain duplicates")
}

func Test_NoAuthority_isSorted(t *testing.T) {
	t.Parallel()

	assert.True(t, slices.IsSorted(hqgourlschemes.NoAuthority), "NoAuthority must be sorted ascending")
}

func Test_NoAuthority_includesData(t *testing.T) {
	t.Parallel()

	_, found := slices.BinarySearch(hqgourlschemes.NoAuthority, "data")

	assert.True(t, found, `"data" must be listed as a no-authority scheme`)
}
