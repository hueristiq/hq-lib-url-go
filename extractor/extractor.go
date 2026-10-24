package extractor

import (
	"fmt"
	"regexp"

	"github.com/hueristiq/hq-lib-url-go/schemes"
)

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
//   - withoutEmail (bool): whether bare email addresses are excluded from extraction.
//   - withoutUserInfo (bool): whether user info in the authority is excluded from extraction.
type Extractor struct {
	withScheme        bool
	withSchemePattern string
	withHost          bool
	withHostPattern   string
	withoutEmail      bool
	withoutUserInfo   bool
}

// CompileRegex constructs and compiles a regular expression pattern for URL extraction.
// It builds a composite regex pattern that can capture various URL forms, taking into account user
// configuration for required schemes or hosts, as well as any custom patterns provided.
// The generated pattern can match fully-qualified URLs (with schemes and hosts), email addresses,
// and relative URLs, depending on the configuration.
//
// Parameters:
//
//	none.
//
// Returns:
//   - regex (*regexp.Regexp): A compiled regular expression object, tuned to find the longest
//     match at each position, that can be used to extract URLs from text.
//   - err (error): a wrapped error when the composed pattern fails to compile, e.g. when a
//     custom scheme or host pattern is invalid.
//
// Example:
//
//	// Create an extractor that requires URL schemes.
//	extractor := New(WithScheme())
//
//	regex, err := extractor.CompileRegex()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	urls := regex.FindAllString(text, -1) // Extracts URLs from the provided text.
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
	noAuthoritySchemePattern := `(?:(?i)` + anyOf(schemes.NoAuthority...) + `:)`

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

// SetScheme sets the Extractor to require URL schemes during extraction.
// It marks the withScheme flag as true.
//
// Parameters:
//
//	none.
//
// Returns:
//
//	none.
func (e *Extractor) SetScheme() {
	e.withScheme = true
}

// SetSchemePattern sets a custom regular expression pattern for matching URL schemes.
// It also marks the Extractor as requiring a scheme.
//
// Parameters:
//   - pattern (string): a regex pattern to match URL schemes.
//
// Returns:
//
//	none.
func (e *Extractor) SetSchemePattern(pattern string) {
	e.SetScheme()

	e.withSchemePattern = pattern
}

// SetHost sets the Extractor to require URL hosts during extraction.
// It marks the withHost flag as true.
//
// Parameters:
//
//	none.
//
// Returns:
//
//	none.
func (e *Extractor) SetHost() {
	e.withHost = true
}

// SetHostPattern sets a custom regular expression pattern for matching URL hosts.
// It also marks the Extractor as requiring a host.
//
// Parameters:
//   - pattern (string): a regex pattern to match URL hosts.
//
// Returns:
//
//	none.
func (e *Extractor) SetHostPattern(pattern string) {
	e.SetHost()

	e.withHostPattern = pattern
}

// SetWithoutEmail marks the Extractor to exclude the bare email address
// pattern from extraction. Email addresses matched through the user info +
// host URL form are unaffected; combine with SetWithoutUserInfo to exclude
// both.
//
// Parameters:
//
//	none.
//
// Returns:
//
//	none.
func (e *Extractor) SetWithoutEmail() {
	e.withoutEmail = true
}

// SetWithoutUserInfo marks the Extractor to exclude user info from the
// authority component of extracted URLs. URLs carrying user info degrade to
// their host (and path); bare email addresses are unaffected, as they match
// through the email pattern.
//
// Parameters:
//
//	none.
//
// Returns:
//
//	none.
func (e *Extractor) SetWithoutUserInfo() {
	e.withoutUserInfo = true
}

// OptionFunc defines a functional option for configuring an Extractor instance.
// It allows the caller to pass in configuration functions that modify the Extractor's
// settings (e.g., requiring a scheme or host, or providing custom regex patterns).
//
// Parameters:
//   - extractor (*Extractor): the Extractor instance to configure.
//
// Returns:
//
//	none.
type OptionFunc func(extractor *Extractor)

// Interface defines the behavior that an Extractor must implement.
type Interface interface {
	// CompileRegex constructs and compiles the regular expression pattern for URL
	// extraction.
	//
	// Parameters:
	//
	//	none.
	//
	// Returns:
	//   - regex (*regexp.Regexp): the compiled pattern, tuned for longest matches.
	//   - err (error): non-nil when the composed pattern is invalid.
	CompileRegex() (regex *regexp.Regexp, err error)
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

	return extractor
}

// WithScheme returns an OptionFunc function that configures the Extractor to require URL schemes in the extraction process.
//
// Parameters:
//
//	none.
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
// Parameters:
//
//	none.
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

// WithoutEmail returns an OptionFunc function that configures the Extractor
// to exclude the bare email address pattern from extraction. Email addresses
// matched through the user info + host URL form are unaffected; combine with
// WithoutUserInfo to exclude both.
//
// Parameters:
//
//	none.
//
// Returns:
//   - f (OptionFunc): a function that marks the withoutEmail flag as true.
func WithoutEmail() (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.SetWithoutEmail()
	}
}

// WithoutUserInfo returns an OptionFunc function that configures the
// Extractor to exclude user info from the authority component of extracted
// URLs. URLs carrying user info degrade to their host (and path); bare email
// addresses are unaffected, as they match through the email pattern.
//
// Parameters:
//
//	none.
//
// Returns:
//   - f (OptionFunc): a function that marks the withoutUserInfo flag as true.
func WithoutUserInfo() (f OptionFunc) {
	return func(extractor *Extractor) {
		extractor.SetWithoutUserInfo()
	}
}
