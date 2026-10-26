package parser

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/hueristiq/hq-lib-url-go/psl"
)

// Parser is responsible for converting raw URL strings into the URL struct that includes
// both the standard URL components and the decomposed domain details. It supports applying
// a default scheme to URLs that lack one, and matches the host against the embedded
// Public Suffix List.
//
// Fields:
//   - scheme (string): the default scheme (e.g., "http", "https") applied when missing.
//   - lenient (bool): whether invalid percent escapes are repaired before parsing.
//   - rules (*psl.List): the Public Suffix List rules used to decompose hosts. [New]
//     installs the full embedded list, [WithIgnorePrivate] installs the ICANN-only list,
//     and a nil value (the zero-value Parser) decomposes hosts positionally.
type Parser struct {
	scheme  string
	lenient bool

	rules *psl.List
}

// SetDefaultScheme sets the default scheme for the Parser. This scheme is applied to URL
// strings that do not already include a scheme.
//
// Parameters:
//   - scheme (string): the default scheme to use (e.g., "http", "https").
func (p *Parser) SetDefaultScheme(scheme string) {
	p.scheme = scheme
}

// Parse converts a raw URL string into a URL struct that encapsulates both the standard
// URL components (parsed by net/url) and the decomposed domain components. If a default
// scheme has been set (via SetDefaultScheme), it is applied to the raw URL string when
// missing. In lenient mode (see [WithLenient]), a failed parse caused by an invalid
// percent escape is retried after [RepairInvalidEscapes] repairs the input.
//
// The host is decomposed into subdomain, SLD, and TLD as follows:
//
//   - An IP literal host (IPv4 or IPv6) yields a nil Domain.
//   - A URL without a host (a relative reference or an opaque URI) yields a non-nil but
//     zero Domain.
//   - A single-label host (e.g. "localhost") lands wholly in SecondLevelDomain, even if
//     the label happens to be a known TLD.
//   - The default rule set is the embedded Public Suffix List: ICANN and private
//     sections, wildcard rules (`*.ck`), exception rules (`!www.ck`), and A-labels.
//     [WithIgnorePrivate] restricts it to the ICANN section. The most specific matching
//     rule determines TopLevelDomain; the label before it lands in SecondLevelDomain, and
//     the remaining labels in Subdomain.
//   - When the parser has no rule set (the zero value), the host is decomposed
//     positionally — every label before the last two lands in Subdomain, the second to
//     last label lands in SecondLevelDomain — and TopLevelDomain is left empty.
//   - When the entire host is a public suffix (e.g. "co.uk"), it lands wholly in
//     TopLevelDomain.
//   - An unknown TLD under the default Public Suffix List uses the implicit "*" rule,
//     so the last label becomes TopLevelDomain, matching the reference algorithm.
//
// Matching is case-insensitive: a trailing root dot is trimmed and the Domain components
// are stored lowercase, regardless of the host's casing.
//
// Parameters:
//   - raw (string): the raw URL string to parse.
//
// Returns:
//   - parsed (*URL): the resulting URL with embedded net/url.URL and domain details.
//   - err (error): an error if the URL cannot be parsed.
func (p *Parser) Parse(raw string) (parsed *URL, err error) {
	if p.scheme != "" {
		raw = p.addScheme(raw)
	}

	u, err := url.Parse(raw)
	if err != nil && p.lenient {
		if repaired := RepairInvalidEscapes(raw); repaired != raw {
			u, err = url.Parse(repaired)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("parser.Parse: %w", err)
	}

	return &URL{
		URL:    u,
		Domain: p.parseDomain(u.Hostname()),
	}, nil
}

// addScheme is a helper method that ensures the URL string is absolute by adding the
// configured default scheme if one is missing. A ":" in the first path segment (before
// any "/") marks an absolute URI that already carries an explicit scheme — possibly an
// opaque one such as "mailto:user@example.com" — so such input is returned unchanged.
// The exception is a scheme-less "host:port" input (digits after the colon), which is
// a host and gets the default scheme.
//
// Parameters:
//   - raw (string): the input URL string which may be missing a scheme.
//
// Returns:
//   - absolute (string): the URL string with the scheme added if necessary.
func (p *Parser) addScheme(raw string) (absolute string) {
	switch {
	case strings.HasPrefix(raw, "//"):
		absolute = p.scheme + ":" + raw
	case strings.HasPrefix(raw, "://"):
		absolute = p.scheme + raw
	case !strings.Contains(raw, "//"):
		if firstSegment, _, _ := strings.Cut(raw, "/"); strings.Contains(firstSegment, ":") && !hasNumericPort(firstSegment) {
			return raw
		}

		absolute = p.scheme + "://" + raw
	default:
		absolute = raw
	}

	return absolute
}

// parseDomain decomposes a hostname into a Domain following the rules documented on
// [Parser.Parse]. The hostname is expected without a port; IP literals yield nil.
//
// Parameters:
//   - hostname (string): the host to decompose, without a port.
//
// Returns:
//   - domain (*Domain): the decomposed domain, or nil for IP literal hosts.
func (p *Parser) parseDomain(hostname string) (domain *Domain) {
	if hostname == "" {
		return &Domain{}
	}

	if net.ParseIP(hostname) != nil {
		return nil
	}

	hostname = strings.TrimSuffix(strings.ToLower(hostname), ".")

	parts := strings.Split(hostname, ".")

	if len(parts) == 1 {
		return &Domain{SecondLevelDomain: hostname}
	}

	rule := p.rules.Find(hostname)
	if rule == nil {
		// A nil rule list (the zero-value Parser) has no default rule: the last label
		// sits in the TLD position but is not a known suffix, so decompose positionally
		// and leave TopLevelDomain empty.
		return &Domain{
			Subdomain:         strings.Join(parts[:len(parts)-2], "."),
			SecondLevelDomain: parts[len(parts)-2],
		}
	}

	left, suffix := rule.Decompose(hostname)
	if left == "" {
		// The entire host is a public suffix (e.g. "co.uk").
		return &Domain{TopLevelDomain: hostname}
	}

	i := strings.LastIndexByte(left, '.')
	if i < 0 {
		return &Domain{
			SecondLevelDomain: left,
			TopLevelDomain:    suffix,
		}
	}

	return &Domain{
		Subdomain:         left[:i],
		SecondLevelDomain: left[i+1:],
		TopLevelDomain:    suffix,
	}
}

// OptionFunc defines a function type used for configuring a Parser instance. Options allow
// customization of the Parser (e.g., setting a default scheme or custom TLDs) during
// initialization. When the same option is supplied more than once, the last one wins.
// A nil OptionFunc is ignored. The With* helpers cover common settings, and callers may
// write custom options because [Parser] is exported.
//
// Parameters:
//   - parser (*Parser): the Parser instance to configure.
type OptionFunc func(parser *Parser)

// Interface defines the standard interface for URL parsing functionality.
//
// Methods:
//   - Parse(raw string) (parsed *URL, err error): converts a raw URL string into a
//     parsed [URL].
type Interface interface {
	// Parse converts a raw URL string into a parsed URL struct.
	//
	// Parameters:
	//   - raw (string): the raw URL string to parse.
	//
	// Returns:
	//   - parsed (*URL): the parsed URL with domain details.
	//   - err (error): non-nil when the URL cannot be parsed.
	Parse(raw string) (parsed *URL, err error)
}

// Compile-time guard ensuring [Parser] satisfies [Interface].
var _ Interface = (*Parser)(nil)

// New creates and initializes a new Parser instance with default settings: no default
// scheme, and the full embedded Public Suffix List (ICANN and private sections) as the
// rule set. [WithIgnorePrivate] restricts it to the ICANN section. Additional
// configurations can be applied via the provided OptionFunc functions.
//
// Parameters:
//   - ofs (...OptionFunc): a variadic list of OptionFunc functions to configure the Parser.
//
// Returns:
//   - parser (*Parser): the newly created Parser instance.
func New(ofs ...OptionFunc) (parser *Parser) {
	parser = &Parser{rules: psl.Default}

	for _, f := range ofs {
		if f != nil {
			f(parser)
		}
	}

	return parser
}

// WithDefaultScheme returns an OptionFunc that sets the default scheme for the Parser.
// This option ensures that URLs missing a scheme are treated as absolute URLs with the
// specified scheme. [InferScheme] can pick a per-host scheme when a fixed one is not
// appropriate.
//
// Parameters:
//   - scheme (string): the default scheme to set (e.g., "http", "https").
//
// Returns:
//   - f (OptionFunc): an OptionFunc function that applies the default scheme to a Parser
//     instance.
func WithDefaultScheme(scheme string) (f OptionFunc) {
	return func(parser *Parser) {
		parser.SetDefaultScheme(scheme)
	}
}

// WithIgnorePrivate returns an OptionFunc that restricts the embedded Public Suffix List
// to its ICANN section, excluding private-domain suffixes such as "blogspot.com" and
// "s3.amazonaws.com". Hosts under a private suffix are then decomposed under the nearest
// ICANN suffix: "foo.blogspot.com" yields SecondLevelDomain "blogspot" and
// TopLevelDomain "com". The implicit "*" rule is preserved, so unknown TLDs still land in
// [Domain.TopLevelDomain].
//
// The default parser includes private suffixes.
//
// Returns:
//   - f (OptionFunc): an OptionFunc function that applies the ICANN-only list.
func WithIgnorePrivate() (f OptionFunc) {
	return func(parser *Parser) {
		parser.rules = psl.ICANN
	}
}

// WithLenient returns an OptionFunc that enables lenient URL parsing: when
// [net/url.Parse] rejects a URL because of an invalid percent escape, the parser
// repairs the escape with [RepairInvalidEscapes] and parses again.
//
// It only repairs percent escapes. It does not relax host validation or accept
// invalid authority characters, so a URL such as "http://exa mple.com" still
// fails. Strict parsing is the default.
//
// Returns:
//   - f (OptionFunc): an OptionFunc function that enables lenient parsing.
func WithLenient() (f OptionFunc) {
	return func(parser *Parser) {
		parser.lenient = true
	}
}
