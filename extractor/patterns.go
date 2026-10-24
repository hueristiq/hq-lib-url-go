package extractor

import (
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/hueristiq/hq-lib-url-go/schemes"
	"github.com/hueristiq/hq-lib-url-go/tlds"
	"github.com/hueristiq/hq-lib-url-go/unicodes"
)

const (
	// alphaCharacterSet defines the set of alphabetical characters.
	alphaCharacterSet = `a-zA-Z`

	// digitCharacterSet defines the set of digit characters.
	digitCharacterSet = `0-9`

	// iUnreservedCharacterSet is the set of unreserved characters allowed in URL
	// components, including letters, digits, hyphens, dots, underscores, tildes, and
	// additional allowed Unicode characters.
	iUnreservedCharacterSet = alphaCharacterSet + digitCharacterSet + `\-\._~` + unicodes.AllowedUcsChar

	// iEndUnreservedCharacterSet is similar to iUnreservedCharacterSet but excludes the
	// dot, so a URL component does not end with punctuation.
	iEndUnreservedCharacterSet = alphaCharacterSet + digitCharacterSet + `\-_~` + unicodes.AllowedUcsCharMinusPunc

	// subDelimsCharacterSet defines the set of sub-delimiter characters allowed in URL
	// components.
	subDelimsCharacterSet = `!\$&'\(\)\*\+,;=`

	// endSubDelimsCharacterSet defines a subset of sub-delimiters typically allowed at
	// the end of URL components.
	endSubDelimsCharacterSet = `\$&\+=`

	// pctEncodingPattern defines the pattern for percent-encoded characters.
	pctEncodingPattern = `%[0-9a-fA-F]{2}`

	// iUserInfoPattern matches user information (e.g., username:password) in a URL. It
	// allows unreserved characters, sub-delimiters, colons, or percent-encoded
	// characters, followed by an "@".
	iUserInfoPattern = `(?:(?:[` + iUnreservedCharacterSet + subDelimsCharacterSet + `:]|` + pctEncodingPattern + `)+@)`

	// iUserInfoOptionalPattern makes the user information optional.
	iUserInfoOptionalPattern = iUserInfoPattern + `?`

	// midIPathSegmentCharacterSet defines the set of characters allowed in the middle of
	// a URL path segment.
	midIPathSegmentCharacterSet = iUnreservedCharacterSet + `%` + subDelimsCharacterSet + `:@`

	// endIPathSegmentCharacterSet defines the set of characters allowed at the end of a
	// URL path segment.
	endIPathSegmentCharacterSet = iEndUnreservedCharacterSet + `%` + endSubDelimsCharacterSet

	// iPrivateCharacters defines Unicode private use areas that are allowed in URLs.
	iPrivateCharacters = `\x{E000}-\x{F8FF}\x{F0000}-\x{FFFFD}\x{100000}-\x{10FFFD}`

	// midICharacterSet is the set of characters allowed in the middle of a URL
	// component.
	midICharacterSet = `/?#\\` + midIPathSegmentCharacterSet + iPrivateCharacters

	// endICharacterSet is the set of characters allowed at the end of a URL component.
	endICharacterSet = `/#` + endIPathSegmentCharacterSet + iPrivateCharacters

	// wellParenPattern, wellBrackPattern, and wellBracePattern match well-formed nested
	// parentheses, brackets, and braces, respectively. wellAllPattern combines these
	// into a single pattern.
	wellParenPattern = `\((?:[` + midICharacterSet + `]|\([` + midICharacterSet + `]*\))*\)`
	wellBrackPattern = `\[(?:[` + midICharacterSet + `]|\[[` + midICharacterSet + `]*\])*\]`
	wellBracePattern = `\{(?:[` + midICharacterSet + `]|\{[` + midICharacterSet + `]*\})*\}`
	wellAllPattern   = wellParenPattern + `|` + wellBrackPattern + `|` + wellBracePattern

	// pathContPattern defines the pattern for a URL path continuation. It matches
	// segments composed of allowed middle characters and ensures that the path does not
	// abruptly end with disallowed characters.
	pathContPattern = `(?:[` + midICharacterSet + `]*(?:` + wellAllPattern + `|[` + endICharacterSet + `]))+`

	// letterPattern, markPattern, and numberPattern are Unicode property patterns for
	// letters, marks, and numbers.
	letterPattern = `\p{L}`
	markPattern   = `\p{M}`
	numberPattern = `\p{N}`

	// iriCharactersPattern defines a pattern for a valid IRI character sequence. It
	// matches a character with Unicode letter, mark, or number properties, possibly
	// followed by a hyphenated sequence.
	iriCharactersPattern = `[` + letterPattern + markPattern + numberPattern + `](?:[` + letterPattern + markPattern + numberPattern + `\-]*[` + letterPattern + markPattern + numberPattern + `])?`

	// subdomainPattern matches one or more IRI character sequences followed by a dot,
	// forming a subdomain.
	subdomainPattern = `(?:` + iriCharactersPattern + `\.)+`

	// punycodePattern matches punycode labels used for internationalized domain names.
	punycodePattern = `xn--[a-z0-9-]+`
)

var (
	// The exported patterns below are the building blocks of the composed extraction
	// pattern. They are mutable variables only so they can be spliced into larger
	// pattern strings; callers must not modify them.
	//
	// SchemePattern defines a general pattern for matching URL schemes.
	// It matches any URL scheme that starts with an alphabetical character (a-z, A-Z), followed by
	// any combination of letters, digits, dots (.), hyphens (-), or plus signs (+), and ends with "://".
	// Additionally, it matches schemes from a predefined list that do not require an authority (host),
	// ending with just a colon (":"). These are known as "no-authority" schemes (e.g., "mailto:").
	//
	// This pattern covers a broad range of schemes, making it versatile for extracting different types
	// of URLs, whether they require an authority component or not.
	SchemePattern = `(?:[a-zA-Z][a-zA-Z0-9.\-+]*://|(?:(?i)` + anyOf(schemes.NoAuthority...) + `:))`

	// KnownOfficialSchemePattern defines a pattern for matching officially recognized
	// URL schemes. These include well-known schemes such as "http", "https", "ftp", etc., as registered
	// with IANA. The pattern ensures that the scheme is followed by "://".
	//
	// This pattern ensures that only officially recognized schemes are matched. Matching is
	// case-insensitive, as required by RFC 3986.
	KnownOfficialSchemePattern = `(?:(?i)` + anyOf(schemes.Official...) + `://)`

	// KnownUnofficialSchemePattern defines a pattern for matching unofficial or less commonly
	// used URL schemes. These schemes may not be registered with IANA but are still valid in specific contexts,
	// such as application-specific schemes (e.g., "slack://", "zoommtg://").
	// The pattern ensures that the scheme is followed by "://".
	//
	// This pattern is useful for applications that work with unofficial or niche schemes.
	// Matching is case-insensitive, as required by RFC 3986.
	KnownUnofficialSchemePattern = `(?:(?i)` + anyOf(schemes.Unofficial...) + `://)`

	// KnownNoAuthoritySchemePattern defines a pattern for matching URL schemes that
	// do not require an authority component (host). These schemes are followed by a colon (":") rather than "://".
	// Examples include "mailto:", "tel:", and "sms:".
	//
	// This pattern is used for schemes where a host is not applicable, making it suitable for schemes
	// that involve direct communication (e.g., email or telephone). Matching is case-insensitive, as
	// required by RFC 3986.
	KnownNoAuthoritySchemePattern = `(?:(?i)` + anyOf(schemes.NoAuthority...) + `:)`

	// KnownSchemePattern combines the patterns for officially recognized, unofficial,
	// and no-authority-required schemes into a single comprehensive pattern.
	// It is case-insensitive (denoted by "(?i)") and matches the broadest possible range of URLs.
	//
	// This pattern is suitable for extracting any known scheme, regardless of its official status
	// or whether it requires an authority component.
	KnownSchemePattern = `(?:(?i)(?:` + anyOf(schemes.Official...) + `|` + anyOf(schemes.Unofficial...) + `)://|` + anyOf(schemes.NoAuthority...) + `:)`

	// IPv4Pattern defines a pattern for matching valid IPv4 addresses.
	// It matches four groups of 1 to 3 digits (0-255) separated by periods (e.g., "192.168.0.1").
	//
	// This pattern is essential for extracting or validating IPv4 addresses in URLs or hostnames.
	IPv4Pattern = `(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])\.(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])\.(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])\.(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])`

	// NonEmptyIPv6Pattern defines a detailed pattern for matching valid, non-empty IPv6 addresses.
	// It accounts for various valid formats of IPv6 addresses, including those with elisions ("::") and IPv4
	// address representations.
	//
	// This pattern supports matching fully expanded IPv6 addresses, elided sections, and IPv4-mapped IPv6 addresses.
	NonEmptyIPv6Pattern = `(?:` +
		// 7 colon-terminated chomps, followed by a final chomp or the rest of an elision.
		`(?:[0-9a-fA-F]{1,4}:){7}(?:[0-9a-fA-F]{1,4}|:)|` +
		// 6 chomps, followed by an IPv4 address or elision with final chomp or final elision.
		`(?:[0-9a-fA-F]{1,4}:){6}(?:` + IPv4Pattern + `|:[0-9a-fA-F]{1,4}|:)|` +
		// 5 chomps, followed by an elision with optional IPv4 or up to 2 final chomps.
		`(?:[0-9a-fA-F]{1,4}:){5}(?::` + IPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,2}|:)|` +
		// 4 chomps, followed by an elision with optional IPv4 (optionally preceded by a chomp) or
		// up to 3 final chomps.
		`(?:[0-9a-fA-F]{1,4}:){4}(?:(?::[0-9a-fA-F]{1,4}){0,1}:` + IPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,3}|:)|` +
		// 3 chomps, followed by an elision with optional IPv4 (preceded by up to 2 chomps) or
		// up to 4 final chomps.
		`(?:[0-9a-fA-F]{1,4}:){3}(?:(?::[0-9a-fA-F]{1,4}){0,2}:` + IPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,4}|:)|` +
		// 2 chomps, followed by an elision with optional IPv4 (preceded by up to 3 chomps) or
		// up to 5 final chomps.
		`(?:[0-9a-fA-F]{1,4}:){2}(?:(?::[0-9a-fA-F]{1,4}){0,3}:` + IPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,5}|:)|` +
		// 1 chomp, followed by an elision with optional IPv4 (preceded by up to 4 chomps) or
		// up to 6 final chomps.
		`(?:[0-9a-fA-F]{1,4}:){1}(?:(?::[0-9a-fA-F]{1,4}){0,4}:` + IPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,6}|:)|` +
		// elision, followed by optional IPv4 (preceded by up to 5 chomps) or up to 7 final chomps.
		// `:` is an intentionally omitted alternative, to avoid matching `::`.
		`:(?:(?::[0-9a-fA-F]{1,4}){0,5}:` + IPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,7})` +
		`)`

	// IPv6Pattern is a comprehensive pattern that matches both fully expanded and compressed IPv6 addresses.
	// It also handles "::" elision and optional IPv4-mapped sections.
	IPv6Pattern = `(?:` + NonEmptyIPv6Pattern + `|::)`

	// PortPattern defines a pattern for matching port numbers in URLs. It
	// matches a colon (":") followed by a valid port number (0 to 65535), as typically
	// found in network addresses. The trailing word boundary keeps an oversized port
	// (e.g. ":65536") from matching on its leading digits.
	PortPattern = `(?::(?:[0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])\b)`

	// PortOptionalPattern is similar to PortPattern but makes the port number optional.
	// This is useful for matching URLs where the port may or may not be specified.
	PortOptionalPattern = PortPattern + `?`
)

// boundedTLDs and unboundedTLDs partition tlds.Official and tlds.Pseudo once at
// startup by whether a TLD ends in an ASCII word character: boundedTLDs are safe to
// follow with a "\b" word boundary in the compiled regex, unboundedTLDs are not
// (Go's "\b" only recognizes ASCII word characters, so a trailing boundary after a
// non-ASCII TLD cannot be relied on). The split is computed entry by entry, so it
// never depends on the generated list's sort order.
var boundedTLDs, unboundedTLDs = splitTLDs(slices.Concat(tlds.Official, tlds.Pseudo))

// knownTLDPatternOnce memoizes the known-TLD alternation, which splices the bounded and
// unbounded TLD lists into a large pattern string. Composing it is done once per process
// (on first use) instead of on every CompileRegex call.
var knownTLDPatternOnce = sync.OnceValue(func() string {
	// Combines punycode, bounded TLDs (followed by a word boundary), and unbounded
	// TLDs into a single pattern, kept case-insensitive with `(?i)`.
	return `(?:(?i)` + punycodePattern + `|` + anyOf(boundedTLDs...) + `\b|` + anyOf(unboundedTLDs...) + `)`
})

// anyOf is a helper function that constructs a non-capturing regex pattern from a list of strings.
// It joins the provided strings with a "|" (alternation) operator and escapes each string to ensure
// special regex characters are treated literally.
//
// Parameters:
//   - strs (...string): a variadic list of strings to be included in the pattern.
//
// Returns:
//   - pattern (string): a regex pattern that matches any one of the provided strings.
func anyOf(strs ...string) (pattern string) {
	var b strings.Builder

	b.WriteString("(?:")

	for i, s := range strs {
		if i != 0 {
			b.WriteByte('|')
		}

		b.WriteString(regexp.QuoteMeta(s))
	}

	b.WriteByte(')')

	pattern = b.String()

	return pattern
}

// splitTLDs partitions list into TLDs that end in an ASCII word character and TLDs that
// do not. The split is deterministic: it inspects each entry on its own, so it never
// depends on the list's sort order.
//
// Parameters:
//   - list ([]string): the TLDs to partition.
//
// Returns:
//   - bounded ([]string): TLDs ending in an ASCII word character, safe to follow with a
//     "\b" word boundary.
//   - unbounded ([]string): TLDs ending in a non-word or non-ASCII character.
func splitTLDs(list []string) (bounded, unbounded []string) {
	for _, tld := range list {
		if hasASCIIWordEnding(tld) {
			bounded = append(bounded, tld)
		} else {
			unbounded = append(unbounded, tld)
		}
	}

	return bounded, unbounded
}

// hasASCIIWordEnding reports whether s ends in an ASCII word character ([0-9A-Za-z_]).
//
// Parameters:
//   - s (string): the string to inspect.
//
// Returns:
//   - ok (bool): true when the last byte of s is an ASCII word character.
func hasASCIIWordEnding(s string) (ok bool) {
	if s == "" {
		return false
	}

	c := s[len(s)-1]

	ok = c == '_' ||
		c >= '0' && c <= '9' ||
		c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z'

	return ok
}
