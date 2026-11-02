package parser

import "strings"

// DecodeUnreservedEscapes decodes percent-encoded RFC 3986 unreserved bytes
// (ALPHA / DIGIT / "-" / "_" / "~") in raw and leaves every other escape
// untouched. A dot deliberately stays encoded, as do reserved escapes such as
// "%2F", so the result never gains a new path separator or parent-directory
// segment.
//
// Parameters:
//   - raw (string): the URL or path to decode.
//
// Returns:
//   - decoded (string): the input with unreserved escapes decoded.
func DecodeUnreservedEscapes(raw string) (decoded string) {
	if !strings.Contains(raw, "%") {
		return raw
	}

	var builder strings.Builder

	builder.Grow(len(raw))

	for index := 0; index < len(raw); {
		if index+2 < len(raw) && raw[index] == '%' && isHexDigit(raw[index+1]) && isHexDigit(raw[index+2]) {
			value := hexValue(raw[index+1])<<4 | hexValue(raw[index+2])

			if isUnreserved(value) {
				builder.WriteByte(value)
			} else {
				builder.WriteString(raw[index : index+3])
			}

			index += 3

			continue
		}

		builder.WriteByte(raw[index])

		index++
	}

	decoded = builder.String()

	return
}

// isUnreserved reports whether a byte may be decoded from a percent escape.
// It follows the RFC 3986 unreserved set except for ".", which stays encoded so
// a "%2E%2E" sequence can never become a parent-directory segment.
//
// Parameters:
//   - b (byte): the byte to inspect.
//
// Returns:
//   - unreserved (bool): true for ALPHA, DIGIT, "-", "_", or "~".
func isUnreserved(b byte) (unreserved bool) {
	switch {
	case b >= 'A' && b <= 'Z', b >= 'a' && b <= 'z', b >= '0' && b <= '9':
		unreserved = true
	case b == '-', b == '_', b == '~':
		unreserved = true
	}

	return
}

// hexValue converts one ASCII hexadecimal digit to its numeric value. The
// caller must have validated the digit with [isHexDigit].
//
// Parameters:
//   - b (byte): the hexadecimal digit.
//
// Returns:
//   - value (byte): the numeric value, 0 through 15.
func hexValue(b byte) (value byte) {
	switch {
	case b >= '0' && b <= '9':
		value = b - '0'
	case b >= 'a' && b <= 'f':
		value = b - 'a' + 10
	default:
		value = b - 'A' + 10
	}

	return
}
