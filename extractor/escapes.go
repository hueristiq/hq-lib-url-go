package extractor

import (
	"strings"
)

// maxUnicodeCodePoint is the highest valid Unicode code point.
const maxUnicodeCodePoint = 0x10FFFF

// DecodeJSEscapes converts the escape sequences found in JavaScript and JSON strings
// into the characters they represent. It understands single-character escapes
// (\b \f \n \r \t \v \' \" \\), hex escapes (\xNN), unicode escapes (\uNNNN,
// including surrogate pairs), braced code point escapes (\u{...}), and octal
// escapes (\NNN with JavaScript's digit rules).
//
// An unrecognized escape decodes to the escaped character itself, mirroring
// JavaScript's NonEscapeCharacter rule and JSON's \/ escape. No characters are
// dropped, and a backslash-newline pair decodes to nothing, matching a
// JavaScript line continuation.
//
// DecodeJSEscapes is meant for scanning raw text (script sources, JSON bodies) so that
// escaped URLs become matchable by regular expressions; it is not a complete
// JavaScript string evaluator.
//
// Parameters:
//   - in (string): the text to decode.
//
// Returns:
//   - decoded (string): in with recognized escape sequences expanded.
func DecodeJSEscapes(in string) (decoded string) {
	var out strings.Builder

	out.Grow(len(in))

	for i := 0; i < len(in); {
		if in[i] != '\\' {
			out.WriteByte(in[i])
			i++

			continue
		}

		// A trailing backslash has nothing to escape.
		if i+1 >= len(in) {
			out.WriteByte(in[i])

			break
		}

		next := in[i+1]

		switch {
		case isSingleEscape(next):
			out.WriteByte(singleEscapeValue(next))

			i += 2
		case next == '\n', next == '\r':
			// A line continuation: the backslash and newline produce nothing.
			if next == '\r' && i+2 < len(in) && in[i+2] == '\n' {
				i += 3
			} else {
				i += 2
			}
		case isOctalDigit(next):
			// \0 not followed by another octal digit is a NUL, not an octal
			// escape.
			if next == '0' && (i+2 >= len(in) || !isOctalDigit(in[i+2])) {
				out.WriteByte(0)

				i += 2

				continue
			}

			maxDigits := 3

			if next >= '4' {
				// \400 is a space followed by "0", not 0o400.
				maxDigits = 2
			}

			end := i + 1

			for count := 0; count < maxDigits && end < len(in) && isOctalDigit(in[end]); count++ {
				end++
			}

			out.WriteByte(octalValue(in[i+1 : end]))
			i = end
		case next == 'x':
			if i+4 <= len(in) && isHexDigit(in[i+2]) && isHexDigit(in[i+3]) {
				out.WriteByte(hexDigitValue(in[i+2])<<4 | hexDigitValue(in[i+3]))

				i += 4
			} else {
				out.WriteByte(next)

				i += 2
			}
		case next == 'u':
			i = decodeUnicodeEscape(&out, in, i)
		default:
			// NonEscapeCharacter: \/ \. \d ... decode to the character.
			out.WriteByte(next)

			i += 2
		}
	}

	decoded = out.String()

	return decoded
}

// decodeUnicodeEscape decodes the unicode escape starting at i (in[i] is the
// backslash and in[i+1] the 'u') into out, and returns the index just past the
// escape.
//
// Parameters:
//   - out (*strings.Builder): the destination the decoded rune is written to.
//   - in (string): the source text containing the escape.
//   - i (int): the byte offset of the backslash.
//
// Returns:
//   - next (int): the byte offset just past the decoded escape.
func decodeUnicodeEscape(out *strings.Builder, in string, i int) (next int) {
	// Braced code point escape: \u{...}
	if i+2 < len(in) && in[i+2] == '{' {
		end := i + 3

		for end < len(in) && isHexDigit(in[end]) {
			end++
		}

		if end > i+3 && end < len(in) && in[end] == '}' {
			if value, ok := parseHexWithin(in[i+3:end], maxUnicodeCodePoint); ok {
				//nolint:gosec // G115: parseHexWithin bounds value by maxUnicodeCodePoint.
				out.WriteRune(rune(value))
			}

			return end + 1
		}

		out.WriteByte('u')

		return i + 2
	}

	// Fixed-width escape: \uNNNN
	if i+6 <= len(in) && isHexQuad(in[i+2:i+6]) {
		value := parseHex(in[i+2 : i+6])
		end := i + 6

		// Combine UTF-16 surrogate pairs into a single code point.
		if value >= 0xD800 && value <= 0xDBFF && end+6 <= len(in) && in[end] == '\\' && in[end+1] == 'u' && isHexQuad(in[end+2:end+6]) {
			low := parseHex(in[end+2 : end+6])

			if low >= 0xDC00 && low <= 0xDFFF {
				value = 0x10000 + (value-0xD800)<<10 + (low - 0xDC00)
				end += 6
			}
		}

		if value <= maxUnicodeCodePoint {
			out.WriteRune(rune(value))
		}

		return end
	}

	out.WriteByte('u')

	return i + 2
}

// isSingleEscape reports whether b names a single-character escape.
//
// Parameters:
//   - b (byte): the character following the backslash.
//
// Returns:
//   - single (bool): true when b is one of \b \f \n \r \t \v \' \" \\.
func isSingleEscape(b byte) (single bool) {
	switch b {
	case 'b', 'f', 'n', 'r', 't', 'v', '\'', '"', '\\':
		return true
	}

	return false
}

// singleEscapeValue returns the byte represented by a single-character escape.
//
// Parameters:
//   - b (byte): the character following the backslash; must satisfy
//     [isSingleEscape].
//
// Returns:
//   - value (byte): the decoded character.
func singleEscapeValue(b byte) (value byte) {
	switch b {
	case 'b':
		return '\b'
	case 'f':
		return '\f'
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	case 'v':
		return '\v'
	default:
		return b
	}
}

// isHexQuad reports whether s holds exactly four hexadecimal digits.
//
// Parameters:
//   - s (string): the string to inspect; it must have at least four bytes.
//
// Returns:
//   - quad (bool): true when the first four bytes of s are hexadecimal digits.
func isHexQuad(s string) (quad bool) {
	return isHexDigit(s[0]) && isHexDigit(s[1]) && isHexDigit(s[2]) && isHexDigit(s[3])
}

// isHexDigit reports whether b is an ASCII hexadecimal digit.
//
// Parameters:
//   - b (byte): the byte to test.
//
// Returns:
//   - hex (bool): true for 0-9, a-f, and A-F.
func isHexDigit(b byte) (hex bool) {
	switch {
	case b >= '0' && b <= '9', b >= 'a' && b <= 'f', b >= 'A' && b <= 'F':
		return true
	}

	return false
}

// isOctalDigit reports whether b is an ASCII octal digit.
//
// Parameters:
//   - b (byte): the byte to test.
//
// Returns:
//   - octal (bool): true for 0-7.
func isOctalDigit(b byte) (octal bool) {
	return b >= '0' && b <= '7'
}

// hexDigitValue returns the numeric value of an ASCII hexadecimal digit.
//
// Parameters:
//   - b (byte): the digit to convert; must satisfy [isHexDigit].
//
// Returns:
//   - value (byte): the digit's value, 0 through 15.
func hexDigitValue(b byte) (value byte) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10
	default:
		return b - 'A' + 10
	}
}

// parseHex interprets a string of hexadecimal digits as a number.
//
// Parameters:
//   - s (string): the hexadecimal digits; it must be non-empty and contain only
//     digits accepted by [isHexDigit].
//
// Returns:
//   - value (uint64): the numeric value of s.
func parseHex(s string) (value uint64) {
	for i := range s {
		value = value<<4 | uint64(hexDigitValue(s[i]))
	}

	return value
}

// parseHexWithin interprets a string of hexadecimal digits as a number and
// reports whether the value stays within limit. It stops as soon as the value
// would exceed limit, so arbitrarily long escapes cannot wrap around uint64.
//
// Parameters:
//   - s (string): the hexadecimal digits; it must be non-empty and contain only
//     digits accepted by [isHexDigit].
//   - limit (uint64): the largest acceptable value.
//
// Returns:
//   - value (uint64): the numeric value of s, or zero when it exceeds limit.
//   - ok (bool): true when the value is at most limit.
func parseHexWithin(s string, limit uint64) (value uint64, ok bool) {
	for i := range s {
		// Check before shifting: a value above limit>>4 would overflow past
		// limit even before the next digit is added, and may wrap uint64.
		if value > limit>>4 {
			return 0, false
		}

		value = value<<4 | uint64(hexDigitValue(s[i]))

		if value > limit {
			return 0, false
		}
	}

	return value, true
}

// octalValue interprets a string of octal digits as a number.
//
// Parameters:
//   - s (string): the octal digits; it must contain only digits accepted by
//     [isOctalDigit].
//
// Returns:
//   - value (byte): the numeric value of s.
func octalValue(s string) (value byte) {
	for i := range s {
		value = value<<3 | (s[i] - '0')
	}

	return value
}
