package parser

import (
	"fmt"
	"net"
	"net/url"
	"slices"
	"strings"

	"github.com/hueristiq/hq-go-url/tlds"
)

// URL extends the standard [net/url.URL] struct by embedding it and adding a parsed view of
// the host's domain components.
//
// The embedded *url.URL carries the standard URL components (scheme, user info, host, port,
// path, query, fragment), while the Domain field holds the host decomposed into subdomain,
// second-level domain, and top-level domain. Domain is nil when the host is an IP literal.
//
// This design enables seamless integration with existing HTTP libraries while providing
// enhanced domain parsing functionality.
type URL struct {
	*url.URL

	Domain *Domain
}

// Key returns the URL component identified by key, unfurl-style
// (https://github.com/tomnomnom/unfurl).
//
// The supported keys are:
//
//   - "scheme": the URL scheme (e.g. "https").
//   - "user": the username from the user info (e.g. "user" in "user:pass@host").
//   - "pass": the password from the user info.
//   - "host": the host exactly as written, including the port when present.
//   - "port": the port, without the colon.
//   - "domain": the reconstructed domain (e.g. "www.example.com"); empty for IP hosts.
//   - "subdomain": the subdomain component of the domain (e.g. "www").
//   - "sld": the second-level domain (e.g. "example"); "root" is an alias.
//   - "tld": the recognized top-level domain (e.g. "com" or "co.uk").
//   - "apex": the apex domain — second-level domain plus TLD (e.g. "example.com").
//   - "path": the URL path (e.g. "/users").
//   - "query": the raw query string, without the leading "?".
//   - "fragment": the fragment, without the leading "#".
//
// An unrecognized key, or a component absent from the URL, yields an empty string.
//
// Parameters:
//   - key (string): the component to retrieve.
//
// Returns:
//   - value (string): the component's value, or "" when the key or the component is
//     unknown.
func (u *URL) Key(key string) (value string) {
	if u == nil || u.URL == nil {
		return
	}

	switch key {
	case "scheme":
		value = u.Scheme
	case "user":
		if u.User != nil {
			value = u.User.Username()
		}
	case "pass":
		if u.User != nil {
			value, _ = u.User.Password()
		}
	case "host":
		value = u.Host
	case "port":
		value = u.Port()
	case "domain":
		if u.Domain != nil {
			value = u.Domain.String()
		}
	case "subdomain":
		if u.Domain != nil {
			value = u.Domain.Subdomain
		}
	case "sld", "root":
		if u.Domain != nil {
			value = u.Domain.SecondLevelDomain
		}
	case "tld":
		if u.Domain != nil {
			value = u.Domain.TopLevelDomain
		}
	case "apex":
		if u.Domain != nil {
			value = u.Domain.Apex()
		}
	case "path":
		value = u.Path
	case "query":
		value = u.RawQuery
	case "fragment":
		value = u.Fragment
	}

	return
}

// Normalize rewrites the URL in place into a canonical form, inspired by the
// normalization behavior of urlx (https://github.com/goware/urlx):
//
//   - the scheme is lowercased;
//   - the hostname is lowercased and a single trailing root dot is trimmed;
//   - a port equal to the scheme's default port (e.g. 80 for http, 443 for https) is
//     dropped — a non-default or malformed port is preserved verbatim;
//   - dot segments ("." and "..") are removed from the path per RFC 3986 §5.2.4;
//   - query parameters are sorted by key, when the raw query parses cleanly;
//   - the Domain components are lowercased to match the normalized host.
//
// The fragment and user info are left untouched. Calling Normalize on a zero URL is a
// no-op.
func (u *URL) Normalize() {
	if u == nil || u.URL == nil {
		return
	}

	u.Scheme = strings.ToLower(u.Scheme)

	host, port := splitHostPort(u.Host)

	host = strings.TrimSuffix(strings.ToLower(host), ".")

	if strings.Contains(host, ":") {
		// Restore the brackets around IPv6 literals.
		host = "[" + host + "]"
	}

	if port != "" && port != defaultSchemePorts[u.Scheme] {
		host += ":" + port
	}

	u.Host = host

	u.Path = removeDotSegments(u.Path)
	u.RawPath = ""

	if u.RawQuery != "" {
		if values, err := url.ParseQuery(u.RawQuery); err == nil {
			u.RawQuery = values.Encode()
		}
	}

	if u.Domain != nil {
		u.Domain.Subdomain = strings.ToLower(u.Domain.Subdomain)
		u.Domain.SecondLevelDomain = strings.ToLower(u.Domain.SecondLevelDomain)
		u.Domain.TopLevelDomain = strings.ToLower(u.Domain.TopLevelDomain)
	}
}

// splitHostPort splits a host[:port] authority into host and port substrings. Unlike
// [net/url.URL.Hostname] and [net/url.URL.Port], it is bracket-aware but agnostic to the
// port's validity, so malformed ports are preserved verbatim instead of being absorbed
// into the host. A bare host with more than one colon is treated as an IPv6 literal.
//
// Parameters:
//   - hostport (string): the authority to split, as found in [net/url.URL.Host].
//
// Returns:
//   - host (string): the host, without brackets for IPv6 literals.
//   - port (string): the port, without the colon, or "" when absent.
func splitHostPort(hostport string) (host, port string) {
	if strings.HasPrefix(hostport, "[") {
		if i := strings.LastIndex(hostport, "]"); i >= 0 {
			host = hostport[1:i]

			if rest := hostport[i+1:]; strings.HasPrefix(rest, ":") {
				port = rest[1:]
			}

			return
		}
	}

	if strings.Count(hostport, ":") == 1 {
		host, port, _ = strings.Cut(hostport, ":")

		return
	}

	host = hostport

	return
}

// Domain represents a parsed domain name, broken down into three main components:
//   - Subdomain: The subdomain part of the domain (e.g., "www" in "www.example.com").
//   - SecondLevelDomain (SLD): The core part of the domain (e.g., "example" in "www.example.com").
//   - TopLevelDomain (TLD): The recognized domain extension (e.g., "com" in "www.example.com").
//
// This structure is useful for analysis or manipulation of domain names.
type Domain struct {
	TopLevelDomain    string
	SecondLevelDomain string
	Subdomain         string
}

// String reconstructs a full domain name from its individual components. It joins the non-empty
// components using a dot ("."). Empty components are omitted from the final output.
//
// Examples:
//   - "www.example.com" when all components are provided.
//   - "example.com" when the subdomain is empty.
//   - "example" when only the SLD is provided.
//
// Returns:
//   - domain (string): the reconstructed domain.
func (d *Domain) String() (domain string) {
	parts := make([]string, 0, 3)

	if d.Subdomain != "" {
		parts = append(parts, d.Subdomain)
	}

	if d.SecondLevelDomain != "" {
		parts = append(parts, d.SecondLevelDomain)
	}

	if d.TopLevelDomain != "" {
		parts = append(parts, d.TopLevelDomain)
	}

	domain = strings.Join(parts, ".")

	return
}

// Apex returns the apex (registrable) domain: the second-level domain joined with the
// top-level domain (e.g. "example.com" for "www.example.com"). When one of the two
// components is empty, Apex returns the other.
//
// Returns:
//   - apex (string): the apex domain.
func (d *Domain) Apex() (apex string) {
	switch {
	case d.SecondLevelDomain == "":
		apex = d.TopLevelDomain
	case d.TopLevelDomain == "":
		apex = d.SecondLevelDomain
	default:
		apex = d.SecondLevelDomain + "." + d.TopLevelDomain
	}

	return
}

// Parser is responsible for converting raw URL strings into the URL struct that includes
// both the standard URL components and the decomposed domain details. It supports applying
// a default scheme to URLs that lack one, and matches the host against a sorted list of
// recognized TLDs using binary search.
//
// Fields:
//   - scheme (string): the default scheme (e.g., "http", "https") applied when missing.
//   - tlds ([]string): the recognized TLDs and effective TLDs, kept sorted ascending so
//     lookups can use binary search. Entries are stored lowercase.
type Parser struct {
	scheme string

	tlds []string
}

// SetDefaultScheme sets the default scheme for the Parser. This scheme is applied to URL
// strings that do not already include a scheme.
//
// Parameters:
//   - scheme (string): the default scheme to use (e.g., "http", "https").
func (p *Parser) SetDefaultScheme(scheme string) {
	p.scheme = scheme
}

// SetTLDs configures the Parser to use a custom set of TLDs, replacing any previously
// configured set. The list is lowercased, sorted, and deduplicated, so lookups never
// depend on the order the caller supplies.
//
// Parameters:
//   - list (...string): the TLDs and effective TLDs to recognize.
func (p *Parser) SetTLDs(list ...string) {
	p.tlds = slices.Clone(list)

	for i := range p.tlds {
		p.tlds[i] = strings.ToLower(p.tlds[i])
	}

	slices.Sort(p.tlds)

	p.tlds = slices.Compact(p.tlds)
}

// Parse converts a raw URL string into a URL struct that encapsulates both the standard
// URL components (parsed by net/url) and the decomposed domain components. If a default
// scheme has been set (via SetDefaultScheme), it is applied to the raw URL string when
// missing.
//
// The host is decomposed into subdomain, SLD, and TLD as follows:
//
//   - An IP literal host (IPv4 or IPv6) yields a nil Domain.
//   - A URL without a host (a relative reference or an opaque URI) yields a non-nil but
//     zero Domain.
//   - A single-label host (e.g. "localhost") lands wholly in SecondLevelDomain, even if
//     the label happens to be a known TLD.
//   - When no suffix of the host matches a recognized TLD, the host is decomposed
//     positionally — every label before the last two lands in Subdomain, the second to
//     last label lands in SecondLevelDomain — and TopLevelDomain is left empty: the last
//     label occupies the TLD position but is not recognized.
//   - When the entire host is a recognized suffix (e.g. "co.uk"), it lands wholly in
//     TopLevelDomain.
//   - Otherwise the longest recognized suffix lands in TopLevelDomain, the label before
//     it in SecondLevelDomain, and the remaining labels in Subdomain.
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
	if err != nil {
		err = fmt.Errorf("parser.Parse: %w", err)

		return
	}

	parsed = &URL{
		URL:    u,
		Domain: p.parseDomain(u.Hostname()),
	}

	return
}

// addScheme is a helper method that ensures the URL string is absolute by adding the
// configured default scheme if one is missing. It checks for various URL patterns and
// prepends the scheme accordingly.
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
		absolute = p.scheme + "://" + raw
	default:
		absolute = raw
	}

	return
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

	offset, found := p.findTLDOffset(parts)

	switch {
	case !found:
		// No recognized TLD: the last label sits in the TLD position but is not a known
		// suffix, so decompose the host positionally and leave TopLevelDomain empty.
		domain = &Domain{
			Subdomain:         strings.Join(parts[:len(parts)-2], "."),
			SecondLevelDomain: parts[len(parts)-2],
		}
	case offset < 0:
		// The entire host is a recognized suffix (e.g. "co.uk").
		domain = &Domain{TopLevelDomain: hostname}
	default:
		domain = &Domain{
			Subdomain:         strings.Join(parts[:offset], "."),
			SecondLevelDomain: parts[offset],
			TopLevelDomain:    strings.Join(parts[offset+1:], "."),
		}
	}

	return
}

// findTLDOffset examines the domain components (split by ".") in reverse order to
// determine the starting index of the longest recognized TLD. It binary-searches each
// candidate suffix in the Parser's sorted TLD list, so matching is exact and never
// depends on substring containment.
//
// Parameters:
//   - parts ([]string): the domain labels (e.g., ["www", "example", "com"]).
//
// Returns:
//   - offset (int): the index of the SLD label — one position before the TLD begins — or
//     -1 when the recognized TLD spans the whole host.
//   - found (bool): whether any suffix of the host is a recognized TLD at all.
func (p *Parser) findTLDOffset(parts []string) (offset int, found bool) {
	offset = -1

	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.Join(parts[i:], ".")

		if _, ok := slices.BinarySearch(p.tlds, candidate); !ok {
			break
		}

		offset = i - 1
		found = true
	}

	return
}

// OptionFunc defines a function type used for configuring a Parser instance. Options allow
// customization of the Parser (e.g., setting a default scheme or custom TLDs) during
// initialization. When the same option is supplied more than once, the last one wins.
//
// Parameters:
//   - parser (*Parser): the Parser instance to configure.
type OptionFunc func(parser *Parser)

// Interface defines the standard interface for URL parsing functionality. Any type that
// implements this interface must provide a Parse method to convert a raw URL string into
// a parsed URL struct.
type Interface interface {
	Parse(raw string) (parsed *URL, err error)
}

// Compile-time guard ensuring [Parser] satisfies [Interface].
var _ Interface = (*Parser)(nil)

// defaultSchemePorts maps schemes to the port [URL.Normalize] drops when it appears
// explicitly in the host.
var defaultSchemePorts = map[string]string{
	"ftp":   "21",
	"http":  "80",
	"https": "443",
	"ws":    "80",
	"wss":   "443",
}

// New creates and initializes a new Parser instance with default settings: no default
// scheme, and the union of the official and pseudo TLD lists from the tlds package as
// the recognized TLDs. Additional configurations can be applied via the provided
// OptionFunc functions.
//
// Parameters:
//   - ofs (...OptionFunc): a variadic list of OptionFunc functions to configure the Parser.
//
// Returns:
//   - parser (*Parser): the newly created Parser instance.
func New(ofs ...OptionFunc) (parser *Parser) {
	parser = &Parser{}

	parser.SetTLDs(slices.Concat(tlds.Official, tlds.Pseudo)...)

	for _, f := range ofs {
		f(parser)
	}

	return
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

// WithTLDs returns an OptionFunc that configures the Parser with a custom set of TLDs,
// replacing the default set. This is useful for handling non-standard or niche TLDs that
// may not be included in the default set.
//
// Parameters:
//   - list (...string): the TLDs and effective TLDs to recognize.
//
// Returns:
//   - f (OptionFunc): an OptionFunc function that applies the custom TLDs to the Parser.
func WithTLDs(list ...string) (f OptionFunc) {
	return func(parser *Parser) {
		parser.SetTLDs(list...)
	}
}

// InferScheme guesses a scheme for a host that is missing one, inspired by the scheme
// inference of urlx (https://github.com/goware/urlx).
//
// It returns "http" for local targets — "localhost" and any *.localhost, *.local,
// *.test, or *.invalid name, as well as loopback, private, link-local, and unspecified
// IP literals — and "https" for everything else. A port suffix on host is ignored.
//
// Parameters:
//   - host (string): the host or hostname to infer a scheme for.
//
// Returns:
//   - scheme (string): "http" or "https".
func InferScheme(host string) (scheme string) {
	hostname := host

	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	hostname = strings.TrimSuffix(strings.ToLower(hostname), ".")

	scheme = "https"

	if ip := net.ParseIP(hostname); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
			scheme = "http"
		}

		return
	}

	switch {
	case hostname == "localhost",
		strings.HasSuffix(hostname, ".localhost"),
		strings.HasSuffix(hostname, ".local"),
		strings.HasSuffix(hostname, ".test"),
		strings.HasSuffix(hostname, ".invalid"):
		scheme = "http"
	}

	return
}

// removeDotSegments applies the dot-segment removal algorithm of RFC 3986 §5.2.4 to a
// URL path, resolving "." and ".." segments without touching trailing slashes.
//
// Parameters:
//   - path (string): the path to clean.
//
// Returns:
//   - cleaned (string): the path with dot segments removed.
func removeDotSegments(path string) (cleaned string) {
	if path == "" {
		return
	}

	// pop drops the last segment from an output buffer, everything after its last "/".
	pop := func(output string) string {
		if i := strings.LastIndex(output, "/"); i >= 0 {
			return output[:i]
		}

		return ""
	}

	input := path
	output := ""

	for input != "" {
		switch {
		case strings.HasPrefix(input, "../"):
			input = input[3:]
		case strings.HasPrefix(input, "./"):
			input = input[2:]
		case strings.HasPrefix(input, "/./"):
			input = input[2:]
		case input == "/.":
			input = "/"
		case strings.HasPrefix(input, "/../"):
			input = input[2:]
			output = pop(output)
		case input == "/..":
			input = "/"
			output = pop(output)
		case input == "." || input == "..":
			input = ""
		default:
			// Move the first path segment — a leading "/" plus the characters up to the
			// next "/" or the end of the input — to the output buffer.
			segment := input
			rest := ""

			start := 0

			if strings.HasPrefix(input, "/") {
				start = 1
			}

			if i := strings.Index(input[start:], "/"); i >= 0 {
				segment = input[:start+i]
				rest = input[start+i:]
			}

			output += segment
			input = rest
		}
	}

	cleaned = output

	return
}
