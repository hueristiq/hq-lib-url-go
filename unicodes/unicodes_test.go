package unicodes

import (
	"regexp"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRangesCompileAsCharacterClasses verifies the generated constants can be
// embedded in a regex character class — the exact way the extractor package
// consumes them.
func TestRangesCompileAsCharacterClasses(t *testing.T) {
	t.Parallel()

	require.NotEmpty(t, AllowedUcsChar)
	require.NotEmpty(t, AllowedUcsCharMinusPunc)

	//nolint:gocritic // Generated character ranges; gocritic's badRegexp cannot parse Unicode ranges of this size.
	_, err := regexp.Compile("[" + AllowedUcsChar + "]")
	require.NoError(t, err, "AllowedUcsChar must compile as a character class")

	//nolint:gocritic // Generated character ranges; gocritic's badRegexp cannot parse Unicode ranges of this size.
	_, err = regexp.Compile("[" + AllowedUcsCharMinusPunc + "]")
	require.NoError(t, err, "AllowedUcsCharMinusPunc must compile as a character class")
}

// TestRangesAreAscending verifies that every start-end pair is ordered and
// that pairs do not overlap, so the sets stay disjoint across regeneration.
func TestRangesAreAscending(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]string{
		"AllowedUcsChar":          AllowedUcsChar,
		"AllowedUcsCharMinusPunc": AllowedUcsCharMinusPunc,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var previousEnd rune = -1

			for offset := 0; offset < len(value); {
				start, size := decodeRune(t, value[offset:])
				offset += size

				end := start

				// A '-' delimiter turns the entry into a start-end range.
				if next, nextSize := decodeRune(t, value[offset:]); next == '-' {
					offset += nextSize

					end, size = decodeRune(t, value[offset:])
					offset += size
				}

				assert.LessOrEqual(t, start, end, "range start must not exceed its end")
				assert.Less(t, previousEnd, start, "ranges must not overlap")

				previousEnd = end
			}

			assert.Positive(t, int64(previousEnd), "range string must be non-empty")
		})
	}
}

func decodeRune(t *testing.T, s string) (r rune, size int) {
	t.Helper()

	require.NotEmpty(t, s, "unexpected end of generated range string")

	r, size = utf8.DecodeRuneInString(s)

	require.NotEqual(t, utf8.RuneError, r, "invalid UTF-8 in generated range string")
	require.NotZero(t, size)

	return r, size
}
