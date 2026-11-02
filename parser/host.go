package parser

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/idna"
)

const (
	// MaxHostLength is the maximum length of a normalized host, in bytes.
	MaxHostLength = 253

	// MaxLabelLength is the maximum length of a single host label, in bytes.
	MaxLabelLength = 63
)

// NormalizeHost canonicalizes host for use as a DNS or HTTP hostname: it trims
// surrounding whitespace and a single trailing root dot, converts international
// labels to their ASCII (punycode) form, lowercases the result, and validates the
// length and every label. IP literals are returned in canonical [netip.Addr]
// form.
//
// Parameters:
//   - host (string): the host or IP literal to normalize.
//
// Returns:
//   - normalized (string): the canonical ASCII host.
//   - err (error): non-nil when host is empty, not valid UTF-8, not a valid
//     IDNA name, exceeds [MaxHostLength], or contains an invalid label.
func NormalizeHost(host string) (normalized string, err error) {
	return normalizeHost(host, false)
}

// NormalizeDNSName is [NormalizeHost] for DNS names. It additionally accepts
// underscores in labels, because service labels such as
// "_ldap._tcp.example.com" are common DNS inputs even though they are not valid
// hostnames for HTTP.
//
// Parameters:
//   - name (string): the DNS name or IP literal to normalize.
//
// Returns:
//   - normalized (string): the canonical ASCII name.
//   - err (error): non-nil when the name is invalid.
func NormalizeDNSName(name string) (normalized string, err error) {
	return normalizeHost(name, true)
}

// normalizeHost implements [NormalizeHost] and [NormalizeDNSName].
//
// Parameters:
//   - host (string): the host or IP literal to normalize.
//   - allowUnderscore (bool): whether labels may contain underscores.
//
// Returns:
//   - normalized (string): the canonical ASCII host.
//   - err (error): non-nil when host is invalid.
func normalizeHost(host string, allowUnderscore bool) (normalized string, err error) {
	host = strings.TrimSpace(host)
	host = strings.TrimSuffix(host, ".")

	if host == "" {
		err = errors.New("parser.NormalizeHost: empty host")

		return
	}

	if !utf8.ValidString(host) {
		err = fmt.Errorf("parser.NormalizeHost: host %q is not valid UTF-8", host)

		return
	}

	if address, addressErr := netip.ParseAddr(host); addressErr == nil {
		normalized = address.String()

		return
	}

	labels := strings.Split(host, ".")

	for index, label := range labels {
		if strings.ContainsRune(label, '_') {
			// IDNA rejects underscores even though DNS service labels use
			// them; keep such labels as ASCII and validate them below.
			labels[index] = strings.ToLower(label)

			continue
		}

		var ascii string

		ascii, err = idna.Lookup.ToASCII(label)
		if err != nil {
			err = fmt.Errorf("parser.NormalizeHost: invalid host %q: %w", host, err)

			return
		}

		labels[index] = strings.ToLower(ascii)
	}

	normalized = strings.Join(labels, ".")

	if len(normalized) > MaxHostLength {
		err = fmt.Errorf("parser.NormalizeHost: host %q exceeds %d bytes", host, MaxHostLength)

		return
	}

	for _, label := range strings.Split(normalized, ".") {
		if labelErr := validateLabel(label, allowUnderscore); labelErr != nil {
			err = fmt.Errorf("parser.NormalizeHost: invalid host %q: %w", host, labelErr)

			return
		}
	}

	return
}

// ValidateHost reports whether host is a valid hostname or IP literal according
// to [NormalizeHost]. The normalized form is discarded.
//
// Parameters:
//   - host (string): the host to validate.
//
// Returns:
//   - err (error): non-nil when host is invalid.
func ValidateHost(host string) (err error) {
	_, err = NormalizeHost(host)

	return
}

// validateLabel checks one ASCII host label: non-empty, at most
// [MaxLabelLength] bytes, made of letters, digits, and hyphens (and
// underscores when allowUnderscore is set), and not starting or ending with a
// hyphen.
//
// Parameters:
//   - label (string): the label to validate.
//   - allowUnderscore (bool): whether underscores are accepted.
//
// Returns:
//   - err (error): non-nil when the label is invalid.
func validateLabel(label string, allowUnderscore bool) (err error) {
	if label == "" {
		err = errors.New("empty label")

		return
	}

	if len(label) > MaxLabelLength {
		err = fmt.Errorf("label %q exceeds %d bytes", label, MaxLabelLength)

		return
	}

	for index := range len(label) {
		character := label[index]

		switch {
		case character >= 'a' && character <= 'z':
		case character >= '0' && character <= '9':
		case character == '-':
		case allowUnderscore && character == '_':
		default:
			err = fmt.Errorf("invalid character %q in label %q", character, label)

			return
		}
	}

	if label[0] == '-' || label[len(label)-1] == '-' {
		err = fmt.Errorf("label %q starts or ends with a hyphen", label)

		return
	}

	return
}
