package extractor

import (
	"regexp"
	"slices"
	"strings"

	"github.com/hueristiq/hq-go-url/schemes"
	"github.com/hueristiq/hq-go-url/tlds"
	"github.com/hueristiq/hq-go-url/unicodes"
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
	// ExtractorSchemePattern defines a general pattern for matching URL schemes.
	// It matches any URL scheme that starts with alphabetical characters (a-z, A-Z), followed by
	// any combination of alphabets, dots (.), hyphens (-), or plus signs (+), and ends with "://".
	// Additionally, it matches schemes from a predefined list that do not require an authority (host),
	// ending with just a colon (":"). These are known as "no-authority" schemes (e.g., "mailto:").
	//
	// This pattern covers a broad range of schemes, making it versatile for extracting different types
	// of URLs, whether they require an authority component or not.
	ExtractorSchemePattern = `(?:[a-zA-Z][a-zA-Z.\-+]*://|` + anyOf(schemes.NoAuthority...) + `:)`

	// ExtractorKnownOfficialSchemePattern defines a pattern for matching officially recognized
	// URL schemes. These include well-known schemes such as "http", "https", "ftp", etc., as registered
	// with IANA. The pattern ensures that the scheme is followed by "://".
	//
	// This pattern ensures that only officially recognized schemes are matched.
	ExtractorKnownOfficialSchemePattern = `(?:` + anyOf(schemes.Official...) + `://)`

	// ExtractorKnownUnofficialSchemePattern defines a pattern for matching unofficial or less commonly
	// used URL schemes. These schemes may not be registered with IANA but are still valid in specific contexts,
	// such as application-specific schemes (e.g., "slack://", "zoommtg://").
	// The pattern ensures that the scheme is followed by "://".
	//
	// This pattern is useful for applications that work with unofficial or niche schemes.
	ExtractorKnownUnofficialSchemePattern = `(?:` + anyOf(schemes.Unofficial...) + `://)`

	// ExtractorKnownNoAuthoritySchemePattern defines a pattern for matching URL schemes that
	// do not require an authority component (host). These schemes are followed by a colon (":") rather than "://".
	// Examples include "mailto:", "tel:", and "sms:".
	//
	// This pattern is used for schemes where a host is not applicable, making it suitable for schemes
	// that involve direct communication (e.g., email or telephone).
	ExtractorKnownNoAuthoritySchemePattern = `(?:` + anyOf(schemes.NoAuthority...) + `:)`

	// ExtractorKnownSchemePattern combines the patterns for officially recognized, unofficial,
	// and no-authority-required schemes into a single comprehensive pattern.
	// It is case-insensitive (denoted by "(?i)") and matches the broadest possible range of URLs.
	//
	// This pattern is suitable for extracting any known scheme, regardless of its official status
	// or whether it requires an authority component.
	ExtractorKnownSchemePattern = `(?:(?i)(?:` + anyOf(schemes.Official...) + `|` + anyOf(schemes.Unofficial...) + `)://|` + anyOf(schemes.NoAuthority...) + `:)`

	// ExtractorIPv4Pattern defines a pattern for matching valid IPv4 addresses.
	// It matches four groups of 1 to 3 digits (0-255) separated by periods (e.g., "192.168.0.1").
	//
	// This pattern is essential for extracting or validating IPv4 addresses in URLs or hostnames.
	ExtractorIPv4Pattern = `(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])\.(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])\.(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])\.(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])`

	// ExtractorNonEmptyIPv6Pattern defines a detailed pattern for matching valid, non-empty IPv6 addresses.
	// It accounts for various valid formats of IPv6 addresses, including those with elisions ("::") and IPv4
	// address representations.
	//
	// This pattern supports matching fully expanded IPv6 addresses, elided sections, and IPv4-mapped IPv6 addresses.
	ExtractorNonEmptyIPv6Pattern = `(?:` +
		// 7 colon-terminated chomps, followed by a final chomp or the rest of an elision.
		`(?:[0-9a-fA-F]{1,4}:){7}(?:[0-9a-fA-F]{1,4}|:)|` +
		// 6 chomps, followed by an IPv4 address or elision with final chomp or final elision.
		`(?:[0-9a-fA-F]{1,4}:){6}(?:` + ExtractorIPv4Pattern + `|:[0-9a-fA-F]{1,4}|:)|` +
		// 5 chomps, followed by an elision with optional IPv4 or up to 2 final chomps.
		`(?:[0-9a-fA-F]{1,4}:){5}(?::` + ExtractorIPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,2}|:)|` +
		// 4 chomps, followed by an elision with optional IPv4 (optionally preceded by a chomp) or
		// up to 3 final chomps.
		`(?:[0-9a-fA-F]{1,4}:){4}(?:(?::[0-9a-fA-F]{1,4}){0,1}:` + ExtractorIPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,3}|:)|` +
		// 3 chomps, followed by an elision with optional IPv4 (preceded by up to 2 chomps) or
		// up to 4 final chomps.
		`(?:[0-9a-fA-F]{1,4}:){3}(?:(?::[0-9a-fA-F]{1,4}){0,2}:` + ExtractorIPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,4}|:)|` +
		// 2 chomps, followed by an elision with optional IPv4 (preceded by up to 3 chomps) or
		// up to 5 final chomps.
		`(?:[0-9a-fA-F]{1,4}:){2}(?:(?::[0-9a-fA-F]{1,4}){0,3}:` + ExtractorIPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,5}|:)|` +
		// 1 chomp, followed by an elision with optional IPv4 (preceded by up to 4 chomps) or
		// up to 6 final chomps.
		`(?:[0-9a-fA-F]{1,4}:){1}(?:(?::[0-9a-fA-F]{1,4}){0,4}:` + ExtractorIPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,6}|:)|` +
		// elision, followed by optional IPv4 (preceded by up to 5 chomps) or up to 7 final chomps.
		// `:` is an intentionally omitted alternative, to avoid matching `::`.
		`:(?:(?::[0-9a-fA-F]{1,4}){0,5}:` + ExtractorIPv4Pattern + `|(?::[0-9a-fA-F]{1,4}){1,7})` +
		`)`

	// ExtractorIPv6Pattern is a comprehensive pattern that matches both fully expanded and compressed IPv6 addresses.
	// It also handles "::" elision and optional IPv4-mapped sections.
	ExtractorIPv6Pattern = `(?:` + ExtractorNonEmptyIPv6Pattern + `|::)`

	// ExtractorPortPattern defines a pattern for matching port numbers in URLs. It
	// matches a colon (":") followed by a valid port number (0 to 65535), as typically
	// found in network addresses. The trailing word boundary keeps an oversized port
	// (e.g. ":65536") from matching on its leading digits.
	ExtractorPortPattern = `(?::(?:[0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])\b)`

	// ExtractorPortOptionalPattern is similar to ExtractorPortPattern but makes the port number optional.
	// This is useful for matching URLs where the port may or may not be specified.
	ExtractorPortOptionalPattern = ExtractorPortPattern + `?`
)

// boundedTLDs and unboundedTLDs partition tlds.Official once at startup by whether a TLD
// ends in an ASCII word character: boundedTLDs are safe to follow with a "\b" word
// boundary in the compiled regex, unboundedTLDs are not (Go's "\b" only recognizes ASCII
// word characters, so a boundary after a non-ASCII TLD would never match). The split is
// computed entry by entry, so it never depends on the generated list's sort order.
var boundedTLDs, unboundedTLDs = splitTLDs(tlds.Official)

// Extractor configures the URL extraction process.
// It provides options for controlling whether URL schemes and hosts are mandatory, and allows custom
// regular expression patterns to be specified for these components. This allows fine-grained control over
// the types of URLs that are extracted from text.
//
// Fields:
//   - withScheme (bool): whether a URL scheme (e.g., "http") is required in extracted URLs.
//   - withSchemePattern (string): a custom regular expression pattern to match URL schemes.
//   - withHost (bool): whether a URL host (e.g., domain) is required in extracted URLs.
//   - withHostPattern (string): a custom regular expression pattern to match URL hosts.
type Extractor struct {
	withScheme        bool
	withSchemePattern string
	withHost          bool
	withHostPattern   string
}

// CompileRegex constructs and compiles a regular expression pattern for URL extraction.
// It builds a composite regex pattern that can capture various URL forms, taking into account user
// configuration for required schemes or hosts, as well as any custom patterns provided.
// The generated pattern can match fully-qualified URLs (with schemes and hosts), email addresses,
// and relative URLs, depending on the configuration.
//
// Returns:
//   - regex (*regexp.Regexp): A compiled regular expression object that can be used to extract URLs from text.
//
// Example:
//
//	// Create an extractor that requires URL schemes.
//	extractor := New(WithScheme())
//	regex := extractor.CompileRegex()
//	urls := regex.FindAllString(text, -1) // Extracts URLs from the provided text.
func (e *Extractor) CompileRegex() (regex *regexp.Regexp) {
	schemePattern := ExtractorSchemePattern

	if e.withScheme && e.withSchemePattern != "" {
		schemePattern = e.withSchemePattern
	}

	// knownTLDPattern combines punycode, known ASCII TLDs, pseudo TLDs (like .local, .onion),
	// and the set of unicode TLDs into a single pattern. We add a word boundary (\b) for ASCII TLDs
	// and keep the pattern case-insensitive with `(?i)`.
	knownTLDPattern := `(?:(?i)` + punycodePattern + `|` + anyOf(slices.Concat(boundedTLDs, tlds.Pseudo)...) + `\b|` + anyOf(unboundedTLDs...) + `)`

	// domainPattern matches typical domains by combining subdomains, known TLDs, or 'localhost'.
	domainPattern := `(?:` + subdomainPattern + knownTLDPattern + `|localhost)`

	// Host patterns:
	//   - hostWithoutPortPattern: matches domain, IPv6 (in brackets), or bare IPv4.
	//   - hostWithPortOptionalPattern: hostWithoutPort plus an optional port.
	hostWithoutPortPattern := `(?:` + domainPattern + `|\[` + ExtractorIPv6Pattern + `\]|\b` + ExtractorIPv4Pattern + `\b)`
	hostWithPortOptionalPattern := `(?:` + hostWithoutPortPattern + ExtractorPortOptionalPattern + `)`

	// If a custom host pattern was provided, use it.
	if e.withHost && e.withHostPattern != "" {
		hostWithPortOptionalPattern = e.withHostPattern
	}

	// authorityPattern: user info + host:port.
	authorityPattern := `(?:` + iUserInfoOptionalPattern + hostWithPortOptionalPattern + `)`

	// authorityOptionalPattern allows an optional authority.
	authorityOptionalPattern := authorityPattern + `?`

	// Patterns for the URL forms:
	//   - webURLPattern:          has authority + optional path
	//   - emailPattern:           simplified email pattern, capturing user@host
	//   - urlsWithSchemePattern:  fully supports an optional or mandatory authority, with path
	webURLPattern := authorityPattern + `(?:/` + pathContPattern + `|/)?`
	emailPattern := `(?P<relaxedEmail>[a-zA-Z0-9._%\-+]+@` + hostWithPortOptionalPattern + `)`
	urlsWithSchemePattern := schemePattern + authorityOptionalPattern + pathContPattern

	// If a custom host pattern is provided, ensure we require the authority.
	if e.withHostPattern != "" {
		urlsWithSchemePattern = schemePattern + authorityPattern + `(?:/` + pathContPattern + `|/)?`
	}

	// urlsWithHostPattern tries to match either web-like URLs or emails
	// without a specific scheme.
	urlsWithHostPattern := webURLPattern + `|` + emailPattern

	// relativeURLsPattern captures relative paths (e.g. /path/to/resource or
	// something/like/this).
	relativeURLsPattern := `(\/[\w\/?=&#.-]*)|([\w\/?=&#.-]+?(?:\/[\w\/?=&#.-]+)+)`

	// Build the final combined pattern depending on the configuration:
	var pattern string

	switch {
	case e.withScheme:
		pattern = urlsWithSchemePattern
	case e.withHost:
		pattern = urlsWithSchemePattern + `|` + urlsWithHostPattern
	default:
		pattern = urlsWithSchemePattern + `|` + urlsWithHostPattern + `|` + relativeURLsPattern
	}

	// Compiling the final regex pattern.
	regex = regexp.MustCompile(pattern)

	// Ensures the longest possible match is found.
	regex.Longest()

	return
}

// SetScheme sets the Extractor to require URL schemes during extraction.
// It marks the withScheme flag as true.
func (e *Extractor) SetScheme() {
	e.withScheme = true
}

// SetSchemePattern sets a custom regular expression pattern for matching URL schemes.
// It also marks the Extractor as requiring a scheme.
//
// Parameters:
//   - pattern (string): a regex pattern to match URL schemes.
func (e *Extractor) SetSchemePattern(pattern string) {
	e.SetScheme()

	e.withSchemePattern = pattern
}

// SetHost sets the Extractor to require URL hosts during extraction.
// It marks the withHost flag as true.
func (e *Extractor) SetHost() {
	e.withHost = true
}

// SetHostPattern sets a custom regular expression pattern for matching URL hosts.
// It also marks the Extractor as requiring a host.
//
// Parameters:
//   - pattern (string): a regex pattern to match URL hosts.
func (e *Extractor) SetHostPattern(pattern string) {
	e.SetHost()

	e.withHostPattern = pattern
}

// OptionFunc defines a functional option for configuring an Extractor instance.
// It allows the caller to pass in configuration functions that modify the Extractor's
// settings (e.g., requiring a scheme or host, or providing custom regex patterns).
//
// Parameters:
//   - extractor (*Extractor): the Extractor instance to configure.
type OptionFunc func(extractor *Extractor)

// Interface defines the behavior that an Extractor must implement.
// It requires the implementation of a method to compile a regular expression for URL extraction.
type Interface interface {
	CompileRegex() (regex *regexp.Regexp)
}

// Compile-time guard ensuring [Extractor] satisfies [Interface].
var _ Interface = (*Extractor)(nil)

// New creates a new Extractor instance with optional configuration.
// The options can be used to customize the behavior of the Extractor, such as requiring URL schemes
// or hosts, and specifying custom regex patterns for these components.
//
// Parameters:
//   - ofs (...OptionFunc): a variadic list of OptionFunc functions that modify the Extractor's settings.
//
// Returns:
//   - extractor (*Extractor): the configured Extractor instance.
func New(ofs ...OptionFunc) (extractor *Extractor) {
	extractor = &Extractor{}

	for _, f := range ofs {
		f(extractor)
	}

	return
}

// WithScheme returns an OptionFunc function that configures the Extractor to require URL schemes in the extraction process.
//
// Returns:
//   - f (OptionFunc): a function that sets the withScheme flag to true.
func WithScheme() (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.SetScheme()
	}
}

// WithSchemePattern returns an OptionFunc function that sets a custom regex pattern for matching URL schemes.
//
// Parameters:
//   - pattern (string): a regex pattern to match URL schemes.
//
// Returns:
//   - f (OptionFunc): a function that sets the custom scheme pattern and enables the withScheme flag.
func WithSchemePattern(pattern string) (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.SetSchemePattern(pattern)
	}
}

// WithHost returns an OptionFunc function that configures the Extractor to require URL hosts in the extraction process.
//
// Returns:
//   - f (OptionFunc): a function that sets the withHost flag to true.
func WithHost() (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.SetHost()
	}
}

// WithHostPattern returns an OptionFunc function that sets a custom regex pattern for matching URL hosts.
//
// Parameters:
//   - pattern (string): a regex pattern to match URL hosts.
//
// Returns:
//   - f (OptionFunc): a function that sets the custom host pattern and enables the withHost flag.
func WithHostPattern(pattern string) (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.SetHostPattern(pattern)
	}
}

// Strict returns an OptionFunc that restricts extraction to URLs carrying a scheme,
// inspired by the Strict preset of xurls (https://github.com/mvdan/xurls).
//
// It is equivalent to [WithScheme]: any syntactically valid scheme matches, not just the
// known ones. Combine it with [WithSchemePattern] and [ExtractorKnownSchemePattern] to
// restrict matches to officially recognized schemes instead.
//
// Returns:
//   - f (OptionFunc): a function that configures the Extractor to require schemes.
func Strict() (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.SetScheme()
	}
}

// Relaxed returns an OptionFunc that matches everything [Strict] matches, plus
// scheme-less URLs with a host and bare email addresses, inspired by the Relaxed preset
// of xurls (https://github.com/mvdan/xurls).
//
// It is equivalent to [WithHost]. Email addresses match the "relaxedEmail"
// subexpression, which can be used to filter them as needed.
//
// Returns:
//   - f (OptionFunc): a function that configures the Extractor to require hosts.
func Relaxed() (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.SetHost()
	}
}

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

	return
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

	return
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
		return
	}

	c := s[len(s)-1]

	ok = c == '_' ||
		c >= '0' && c <= '9' ||
		c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z'

	return
}
