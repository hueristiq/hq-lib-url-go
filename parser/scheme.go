package parser

import (
	"net"
	"strings"
)

// defaultSchemePorts maps schemes to the port [URL.Normalize] drops when it appears
// explicitly in the host.
var defaultSchemePorts = map[string]string{
	"ftp":   "21",
	"http":  "80",
	"https": "443",
	"ws":    "80",
	"wss":   "443",
}

// hasNumericPort reports whether segment looks like a "host:port" pair — that is, it
// contains a ":" followed by a non-empty, all-digits port.
//
// Parameters:
//   - segment (string): the first path segment of a scheme-less URL.
//
// Returns:
//   - ok (bool): true when segment carries a numeric port.
func hasNumericPort(segment string) (ok bool) {
	_, port, found := strings.Cut(segment, ":")
	if !found || port == "" {
		return false
	}

	for _, r := range port {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

// InferScheme guesses a scheme for a host that is missing one.
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

		return scheme
	}

	switch {
	case hostname == "localhost",
		strings.HasSuffix(hostname, ".localhost"),
		strings.HasSuffix(hostname, ".local"),
		strings.HasSuffix(hostname, ".test"),
		strings.HasSuffix(hostname, ".invalid"):
		scheme = "http"
	}

	return scheme
}
