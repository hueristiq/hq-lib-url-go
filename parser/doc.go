// Package parser parses raw URL strings into structured values that extend the standard
// net/url package with a decomposition of the host's domain into subdomain, second-level
// domain (SLD), and top-level domain (TLD).
//
// It suits work that needs more from a URL than net/url exposes directly — crawlers,
// reconnaissance tools, and log analyzers that group or filter by registrable domain —
// where embedding [net/url.URL] keeps the result compatible with existing HTTP libraries.
//
// # Parsing
//
// [New] returns a [Parser]; [Parser.Parse] converts a raw URL string into a [URL]. The
// behavior is set through functional options:
//
//   - [WithDefaultScheme] applies a scheme to URLs that lack one, so "example.com/path"
//     parses as a host plus path instead of a relative reference. A URL whose first path
//     segment contains ":" — an opaque URI such as "mailto:user@example.com", or a
//     host:port pair — already carries a scheme per RFC 3986 and is left unchanged.
//   - [WithTLDs] replaces the recognized TLD set — the union of the official and pseudo
//     TLD lists from the tlds package by default.
//
// # Domain decomposition
//
// The host lands in [Domain] as follows:
//
//   - An IP literal host (IPv4 or IPv6) yields a nil Domain.
//   - A URL without a host (a relative reference or an opaque URI) yields a non-nil but
//     zero Domain.
//   - A single-label host (e.g. "localhost") lands wholly in
//     [Domain.SecondLevelDomain], even if the label happens to be a known TLD.
//   - When no suffix of the host is a recognized TLD, the host is decomposed
//     positionally — the labels before the last two land in [Domain.Subdomain], the
//     second to last label lands in [Domain.SecondLevelDomain] — and
//     [Domain.TopLevelDomain] is left empty: the last label occupies the TLD position
//     but is not recognized.
//   - When the entire host is a recognized suffix (e.g. "co.uk"), it lands wholly in
//     [Domain.TopLevelDomain].
//   - Otherwise the longest recognized suffix lands in [Domain.TopLevelDomain], the
//     label before it in [Domain.SecondLevelDomain], and the remaining labels in
//     [Domain.Subdomain].
//
// Matching is case-insensitive and ignores a trailing root dot; the Domain components
// are stored lowercase. TLD lookup is an exact match against a sorted list via binary
// search, so it never depends on the list's source order and never confuses a substring
// of a TLD with the TLD itself. The generated TLD list excludes punycode (xn--) TLDs, so
// hosts ending in one (e.g. "example.xn--p1ai") are not recognized here — while the
// extractor package matches them.
//
// # Key access
//
// [URL.Key] returns a single component by name — "scheme", "user", "pass", "host",
// "port", "domain", "subdomain", "sld" (alias "root"), "tld", "apex", "path", "query",
// or "fragment". [Domain.Apex] returns the apex domain, and [Domain.String]
// reconstructs the full domain.
//
// # Normalization
//
// [URL.Normalize] rewrites a parsed URL into a canonical form: lowercase scheme and
// host, no trailing root dot, no default port, dot segments removed from the path, and
// query parameters sorted by key. [InferScheme] guesses a scheme for a host that is
// missing one — "http" for local targets, "https" otherwise.
//
// The implementation relies on the Go standard library only.
package parser
