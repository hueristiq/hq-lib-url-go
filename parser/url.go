package parser

import (
	"net/url"
	"strings"
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

// Key returns the URL component identified by key.
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
		return ""
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

	return value
}

// Normalize rewrites the URL in place into a canonical form:
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
//
// Parameters:
//
//	none.
//
// Returns:
//
//	none.
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

			return host, port
		}
	}

	if strings.Count(hostport, ":") == 1 {
		host, port, _ = strings.Cut(hostport, ":")

		return host, port
	}

	return hostport, ""
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
		return ""
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
			input = "/" + input[3:]
		case input == "/.":
			input = "/"
		case strings.HasPrefix(input, "/../"):
			input = "/" + input[4:]
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

	return cleaned
}
