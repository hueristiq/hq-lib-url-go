package extractor

import (
	"fmt"
	"regexp"
	"sync"

	hqgourlscheme "github.com/hueristiq/hq-lib-url-go/internal/scheme"
)

// Extractor configures the URL extraction process.
// It provides options for controlling whether URL schemes and hosts are mandatory, and allows custom
// regular expression patterns to be specified for these components. This allows fine-grained control over
// the types of URLs that are extracted from text.
//
// Construct through [New]; an Extractor is immutable after construction and
// must not be copied after first use.
//
// Fields:
//   - withScheme (bool): whether a URL scheme (e.g., "http") is required in extracted URLs.
//   - withSchemePattern (string): a custom regular expression pattern to match URL schemes.
//   - withHost (bool): whether a URL host (e.g., domain) is required in extracted URLs.
//   - withHostPattern (string): a custom regular expression pattern to match URL hosts.
//   - withoutEmail (bool): whether bare email addresses are excluded from extraction.
//   - withoutUserInfo (bool): whether user info in the authority is excluded from extraction.
//   - regexpOnce (sync.Once): guards the lazy compilation of the extraction regexp.
//   - regexp (*regexp.Regexp): the compiled extraction pattern, cached for [Extractor.Extract].
//   - regexpErr (error): the compilation failure cached alongside regexp.
type Extractor struct {
	withScheme        bool
	withSchemePattern string
	withHost          bool
	withHostPattern   string
	withoutEmail      bool
	withoutUserInfo   bool

	regexpOnce sync.Once
	regexp     *regexp.Regexp
	regexpErr  error
}

// CompileRegex constructs and compiles a regular expression pattern for URL extraction.
// It builds a composite regex pattern that can capture various URL forms, taking into account user
// configuration for required schemes or hosts, as well as any custom patterns provided.
// The generated pattern can match fully-qualified URLs (with schemes and hosts), email addresses,
// and relative URLs, depending on the configuration.
//
// Returns:
//   - regex (*regexp.Regexp): a compiled regular expression object, tuned to find the
//     longest match at each position, that can be used to extract URLs from text.
//   - err (error): a wrapped error when the composed pattern fails to compile, e.g. when a
//     custom scheme or host pattern is invalid.
func (e *Extractor) CompileRegex() (regex *regexp.Regexp, err error) {
	knownTLDPattern := knownTLDPatternOnce()

	// domainPattern matches typical domains by combining subdomains, known TLDs, or 'localhost'.
	domainPattern := `(?:` + subdomainPattern + knownTLDPattern + `|localhost)`

	// Host patterns:
	//   - hostWithoutPortPattern: matches domain, IPv6 (in brackets), or bare IPv4.
	//   - hostWithPortOptionalPattern: hostWithoutPort plus an optional port.
	hostWithoutPortPattern := `(?:` + domainPattern + `|\[` + IPv6Pattern + `\]|\b` + IPv4Pattern + `\b)`
	hostWithPortOptionalPattern := `(?:` + hostWithoutPortPattern + PortOptionalPattern + `)`

	// If a custom host pattern was provided, use it. Group it so a top-level
	// alternation cannot change the precedence of the surrounding pattern.
	if e.withHost && e.withHostPattern != "" {
		hostWithPortOptionalPattern = `(?:` + e.withHostPattern + `)`
	}

	// authorityPattern: user info + host:port.
	userInfoPattern := iUserInfoOptionalPattern

	if e.withoutUserInfo {
		userInfoPattern = ""
	}

	authorityPattern := `(?:` + userInfoPattern + hostWithPortOptionalPattern + `)`

	// authorityOptionalPattern allows an optional authority.
	authorityOptionalPattern := authorityPattern + `?`

	// Patterns for the URL forms:
	//   - webURLPattern:          has authority + optional path
	//   - emailPattern:           simplified email pattern, capturing user@host
	//   - urlsWithSchemePattern:  fully supports an optional or mandatory authority, with path
	webURLPattern := authorityPattern + `(?:[/?#]` + pathContPattern + `|/)?`
	emailPattern := `(?P<relaxedEmail>[a-zA-Z0-9._%\-+]+@` + hostWithPortOptionalPattern + `)`

	// The authority and no-authority scheme patterns mirror the exported
	// SchemePattern's components. A custom scheme pattern, when provided,
	// replaces both and is treated as an authority scheme. Group it so a
	// top-level alternation cannot change the precedence of the final pattern.
	authoritySchemePattern := `[a-zA-Z][a-zA-Z0-9.\-+]*://`
	noAuthoritySchemePattern := `(?:(?i)` + anyOf(hqgourlscheme.NoAuthority...) + `:)`

	if e.withScheme && e.withSchemePattern != "" {
		authoritySchemePattern = `(?:` + e.withSchemePattern + `)`
		noAuthoritySchemePattern = ""
	}

	schemePattern := `(?:` + authoritySchemePattern

	if noAuthoritySchemePattern != "" {
		schemePattern += `|` + noAuthoritySchemePattern
	}

	schemePattern += `)`

	// Schemes with an authority take it optionally by default. When user info
	// is excluded, the authority becomes mandatory: leaving it optional would
	// let the path pattern absorb userinfo text ("user@host") as a path.
	// No-authority schemes never take an authority.
	authoritySchemeAuthorityPattern := authorityOptionalPattern
	authoritySchemePathPattern := pathContPattern

	if e.withoutUserInfo {
		authoritySchemeAuthorityPattern = authorityPattern
		authoritySchemePathPattern = `(?:/` + pathContPattern + `|/)?`
	}

	urlsWithSchemePattern := authoritySchemePattern + authoritySchemeAuthorityPattern + authoritySchemePathPattern

	if noAuthoritySchemePattern != "" {
		urlsWithSchemePattern += `|` + noAuthoritySchemePattern + pathContPattern
	}

	// If a custom host pattern is provided, everything must carry the host:
	// scheme URLs and no-authority scheme matches alike.
	if e.withHostPattern != "" {
		urlsWithSchemePattern = schemePattern + authorityPattern + `(?:/` + pathContPattern + `|/)?`
	}

	// urlsWithHostPattern tries to match either web-like URLs or emails
	// without a specific scheme.
	urlsWithHostPattern := webURLPattern

	if !e.withoutEmail {
		urlsWithHostPattern += `|` + emailPattern
	}

	// relativeURLsPattern captures relative paths (e.g. /path/to/resource or
	// something/like/this).
	relativeURLsPattern := `(?:/[\w/?=&#.-]*)|(?:[\w/?=&#.-]+?(?:/[\w/?=&#.-]+)+)`

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
	regex, err = regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("extractor.CompileRegex: %w", err)
	}

	// Ensures the longest possible match is found.
	regex.Longest()

	return regex, nil
}

// Extract finds every URL in text using the configured extraction pattern,
// returning the matches in the order they appear. The pattern is compiled on
// the first call and cached, so repeated calls share one compiled regex.
//
// Parameters:
//   - text (string): the text to search.
//
// Returns:
//   - matches ([]string): the URLs found, in order of appearance; empty when
//     none match.
//   - err (error): a wrapped error when the composed pattern fails to compile,
//     e.g. when a custom scheme or host pattern is invalid.
func (e *Extractor) Extract(text string) (matches []string, err error) {
	e.regexpOnce.Do(func() {
		e.regexp, e.regexpErr = e.CompileRegex()
	})

	if e.regexpErr != nil {
		return nil, e.regexpErr
	}

	matches = e.regexp.FindAllString(text, -1)

	return matches, nil
}

// OptionFunc defines a functional option for configuring an Extractor instance.
// It allows the caller to pass in configuration functions that modify the Extractor's
// settings (e.g., requiring a scheme or host, or providing custom regex patterns).
// A nil OptionFunc is ignored. The With* and Without* helpers cover common settings,
// and callers may write custom options because [Extractor] is exported.
//
// Parameters:
//   - extractor (*Extractor): the Extractor instance to configure.
type OptionFunc func(extractor *Extractor)

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
		if f != nil {
			f(extractor)
		}
	}

	return extractor
}

// WithScheme returns an OptionFunc function that configures the Extractor to require URL schemes in the extraction process.
//
// Returns:
//   - f (OptionFunc): a function that sets the withScheme flag to true.
func WithScheme() (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.withScheme = true
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
		extractor.withScheme = true
		extractor.withSchemePattern = pattern
	}
}

// WithHost returns an OptionFunc function that configures the Extractor to require URL hosts in the extraction process.
//
// Returns:
//   - f (OptionFunc): a function that sets the withHost flag to true.
func WithHost() (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.withHost = true
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
		extractor.withHost = true
		extractor.withHostPattern = pattern
	}
}

// WithoutEmail returns an OptionFunc function that configures the Extractor
// to exclude the bare email address pattern from extraction. Email addresses
// matched through the user info + host URL form are unaffected; combine with
// WithoutUserInfo to exclude both.
//
// Returns:
//   - f (OptionFunc): a function that marks the withoutEmail flag as true.
func WithoutEmail() (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.withoutEmail = true
	}
}

// WithoutUserInfo returns an OptionFunc function that configures the
// Extractor to exclude user info from the authority component of extracted
// URLs. URLs carrying user info degrade to their host (and path); bare email
// addresses are unaffected, as they match through the email pattern.
//
// Returns:
//   - f (OptionFunc): a function that marks the withoutUserInfo flag as true.
func WithoutUserInfo() (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.withoutUserInfo = true
	}
}
