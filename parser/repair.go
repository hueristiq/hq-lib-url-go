package parser

import "strings"

// RepairInvalidEscapes returns raw with every percent sign that does not start a
// valid two-digit hexadecimal escape replaced by "%25". Valid escapes and all
// other bytes are left untouched.
//
// It repairs only percent escapes. Spaces and other invalid authority characters
// are preserved, so [net/url.Parse] can still reject an invalid host.
//
// Parameters:
//   - raw (string): the raw URL string to repair.
//
// Returns:
//   - repaired (string): the repaired URL string.
func RepairInvalidEscapes(raw string) (repaired string) {
	if !strings.Contains(raw, "%") {
		return raw
	}

	var b strings.Builder

	b.Grow(len(raw))

	for i := range len(raw) {
		c := raw[i]
		if c != '%' {
			b.WriteByte(c)

			continue
		}

		if i+2 < len(raw) && isHexDigit(raw[i+1]) && isHexDigit(raw[i+2]) {
			b.WriteByte(c)

			continue
		}

		b.WriteString("%25")
	}

	repaired = b.String()

	return repaired
}

// isHexDigit reports whether c is an ASCII hexadecimal digit.
//
// Parameters:
//   - c (byte): the byte to inspect.
//
// Returns:
//   - ok (bool): true when c is 0-9, a-f, or A-F.
func isHexDigit(c byte) (ok bool) {
	switch {
	case c >= '0' && c <= '9':
		ok = true
	case c >= 'a' && c <= 'f':
		ok = true
	case c >= 'A' && c <= 'F':
		ok = true
	default:
		ok = false
	}

	return ok
}
