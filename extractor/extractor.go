package extractor

import (
	"regexp"

	hqgoerrors "github.com/hueristiq/hq-lib-errors-go"
	hqgourlscheme "github.com/hueristiq/hq-lib-url-go/internal/scheme"
)

// Extractor finds URLs in text with the pattern compiled by [New]. It is
// immutable after construction and safe for concurrent use. Construct through
// [New]; the zero value is not usable.
type Extractor struct {
	regexp *regexp.Regexp
}

// Options configures an [Extractor]. The zero value extracts every supported
// URL form (scheme URLs, host-like URLs, emails, and relative URLs).
type Options struct {
	// RequireScheme requires extracted URLs to carry a scheme (e.g., "http").
	RequireScheme bool
	// SchemePattern is a custom regular expression for the scheme; it implies
	// RequireScheme and replaces both built-in scheme patterns.
	SchemePattern string
	// RequireHost requires extracted URLs to carry a host.
	RequireHost bool
	// HostPattern is a custom regular expression for the host; it implies
	// RequireHost and requires every match to carry the host.
	HostPattern string
	// WithoutEmail excludes bare email addresses from extraction.
	WithoutEmail bool
	// WithoutUserInfo excludes user info from the authority component.
	WithoutUserInfo bool
	// Regexp replaces the composed pattern with a caller-supplied one. It
	// cannot be combined with the other fields.
	Regexp *regexp.Regexp
}

// compileRegex constructs and compiles the extraction pattern from options.
// The pattern can match fully-qualified URLs (with schemes and hosts), email
// addresses, and relative URLs, depending on the configuration. It is tuned
// to find the longest match at each position.
func compileRegex(options Options) (regex *regexp.Regexp, err error) {
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
	if options.RequireHost && options.HostPattern != "" {
		hostWithPortOptionalPattern = `(?:` + options.HostPattern + `)`
	}

	// authorityPattern: user info + host:port.
	userInfoPattern := iUserInfoOptionalPattern

	if options.WithoutUserInfo {
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

	if options.RequireScheme && options.SchemePattern != "" {
		authoritySchemePattern = `(?:` + options.SchemePattern + `)`
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

	if options.WithoutUserInfo {
		authoritySchemeAuthorityPattern = authorityPattern
		authoritySchemePathPattern = `(?:/` + pathContPattern + `|/)?`
	}

	urlsWithSchemePattern := authoritySchemePattern + authoritySchemeAuthorityPattern + authoritySchemePathPattern

	if noAuthoritySchemePattern != "" {
		urlsWithSchemePattern += `|` + noAuthoritySchemePattern + pathContPattern
	}

	// If a custom host pattern is provided, everything must carry the host:
	// scheme URLs and no-authority scheme matches alike.
	if options.HostPattern != "" {
		urlsWithSchemePattern = schemePattern + authorityPattern + `(?:/` + pathContPattern + `|/)?`
	}

	// urlsWithHostPattern tries to match either web-like URLs or emails
	// without a specific scheme.
	urlsWithHostPattern := webURLPattern

	if !options.WithoutEmail {
		urlsWithHostPattern += `|` + emailPattern
	}

	// relativeURLsPattern captures relative paths (e.g. /path/to/resource or
	// something/like/this).
	relativeURLsPattern := `(?:/[\w/?=&#.-]*)|(?:[\w/?=&#.-]+?(?:/[\w/?=&#.-]+)+)`

	// Build the final combined pattern depending on the configuration:
	var pattern string

	switch {
	case options.RequireScheme:
		pattern = urlsWithSchemePattern
	case options.RequireHost:
		pattern = urlsWithSchemePattern + `|` + urlsWithHostPattern
	default:
		pattern = urlsWithSchemePattern + `|` + urlsWithHostPattern + `|` + relativeURLsPattern
	}

	// Compiling the final regex pattern.
	regex, err = regexp.Compile(pattern)
	if err != nil {
		return nil, hqgoerrors.Stack(err)
	}

	// Ensures the longest possible match is found.
	regex.Longest()

	return regex, nil
}

// Extract finds every URL in text using the pattern compiled by [New],
// returning the matches in the order they appear.
func (e *Extractor) Extract(text string) (matches []string) {
	return e.regexp.FindAllString(text, -1)
}

// Regexp returns the compiled extraction pattern for callers that need match
// indexes rather than the strings returned by [Extractor.Extract]. The pattern
// is shared and safe for concurrent matching; callers must not reconfigure it.
func (e *Extractor) Regexp() (regex *regexp.Regexp) {
	return e.regexp
}

// OptionFunc configures an [Options] value. A nil OptionFunc is ignored. The
// With* and Without* helpers cover every setting, and callers may write
// custom options because [Options] is exported.
type OptionFunc func(*Options)

// New creates an Extractor with the pattern described by the options. It
// fails when a pattern is invalid or [WithRegexp] is combined with other
// options.
func New(ofs ...OptionFunc) (extractor *Extractor, err error) {
	options := Options{}

	for _, f := range ofs {
		if f != nil {
			f(&options)
		}
	}

	if options.Regexp != nil {
		if options.RequireScheme || options.RequireHost || options.WithoutEmail || options.WithoutUserInfo {
			return nil, hqgoerrors.New("WithRegexp cannot be combined with other extraction options")
		}

		return &Extractor{regexp: options.Regexp}, nil
	}

	compiled, err := compileRegex(options)
	if err != nil {
		return nil, hqgoerrors.Stack(err)
	}

	return &Extractor{regexp: compiled}, nil
}

// WithRegexp returns an OptionFunc that uses re as the extraction pattern
// instead of building one from the other options. This is useful when the
// caller already has a domain-specific or otherwise custom pattern. The
// pattern is used as given; pair it with [Extractor.ExtractURLs] to keep the
// boundary and glue cleanup behavior. A nil value is ignored.
func WithRegexp(re *regexp.Regexp) (f OptionFunc) {
	return func(o *Options) {
		if re != nil {
			o.Regexp = re
		}
	}
}

// WithScheme returns an OptionFunc that requires extracted URLs to carry a
// scheme.
func WithScheme() (f OptionFunc) {
	return func(o *Options) {
		o.RequireScheme = true
	}
}

// WithSchemePattern returns an OptionFunc that requires extracted URLs to
// carry a scheme matching pattern.
func WithSchemePattern(pattern string) (f OptionFunc) {
	return func(o *Options) {
		o.RequireScheme = true
		o.SchemePattern = pattern
	}
}

// WithHost returns an OptionFunc that requires extracted URLs to carry a
// host.
func WithHost() (f OptionFunc) {
	return func(o *Options) {
		o.RequireHost = true
	}
}

// WithHostPattern returns an OptionFunc that requires extracted URLs to
// carry a host matching pattern.
func WithHostPattern(pattern string) (f OptionFunc) {
	return func(o *Options) {
		o.RequireHost = true
		o.HostPattern = pattern
	}
}

// WithoutEmail returns an OptionFunc that excludes the bare email address
// pattern from extraction. Email addresses matched through the user info +
// host URL form are unaffected; combine with WithoutUserInfo to exclude both.
func WithoutEmail() (f OptionFunc) {
	return func(o *Options) {
		o.WithoutEmail = true
	}
}

// WithoutUserInfo returns an OptionFunc that excludes user info from the
// authority component of extracted URLs. URLs carrying user info degrade to
// their host (and path); bare email addresses are unaffected, as they match
// through the email pattern.
func WithoutUserInfo() (f OptionFunc) {
	return func(o *Options) {
		o.WithoutUserInfo = true
	}
}
