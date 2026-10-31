// Package parser parses raw URL strings into structured values that extend the standard
// net/url package with a decomposition of the host's domain into subdomain, second-level
// domain (SLD), and top-level domain (TLD).
//
// It suits work that needs more from a URL than net/url exposes directly — crawlers,
// reconnaissance tools, and log analyzers that group or filter by registrable domain —
// where embedding [net/url.URL] keeps the result compatible with existing HTTP libraries.
//
// # Entry points
//
// [New] returns a [Parser], and [Parser.Parse] converts a raw URL string into a [URL].
// [RepairInvalidEscapes] repairs invalid percent escapes for callers that parse with
// [net/url] directly.
//
// # Parsing
//
// [Parser.Parse] keeps the standard components — scheme, user info, host, port, path,
// query, and fragment — in the embedded [net/url.URL] and adds the host decomposition
// described below. Host validation and strict percent-escape handling follow net/url;
// [WithLenient] relaxes the latter.
//
// # Domain decomposition
//
// The host lands in [Domain] as follows:
//
//   - An IP literal host (IPv4 or IPv6) yields a nil Domain.
//   - A URL without a host (a relative reference or an opaque URI) yields a non-nil but
//     zero Domain.
//   - A single-label host (e.g. "localhost") lands wholly in
//     [Domain.SecondLevelDomain].
//   - By default the embedded Public Suffix List supplies the rules: its ICANN and
//     private sections, wildcard rules (`*.ck`), exception rules (`!www.ck`), and
//     A-labels. [WithIgnorePrivate] drops the private section. The most specific matching
//     rule determines [Domain.TopLevelDomain]; the label before it lands in
//     [Domain.SecondLevelDomain], and the remaining labels in [Domain.Subdomain].
//   - When the entire host is a public suffix (e.g. "co.uk"), it lands wholly in
//     [Domain.TopLevelDomain].
//   - An unknown TLD under the default rules uses the implicit "*" rule, so the last
//     label becomes [Domain.TopLevelDomain], matching the reference Public Suffix List
//     algorithm.
//   - A zero-value Parser has no rule set and decomposes the host positionally: the
//     labels before the last two land in [Domain.Subdomain], the second to last label
//     lands in [Domain.SecondLevelDomain], and [Domain.TopLevelDomain] is left empty.
//
// Matching is case-insensitive and ignores a trailing root dot; the Domain components
// are stored lowercase. Rule lookup is exact at label boundaries, so a label is never
// confused with a substring of a longer suffix.
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
// # Comparison with weppos/publicsuffix-go
//
// The github.com/weppos/publicsuffix-go/publicsuffix package solves a related
// problem with a domain-only API. The main differences:
//
//   - [Parser.Parse] takes a URL and keeps its scheme, host, path, query, and
//     fragment; weppos takes a bare domain name.
//   - The result here is a [URL] with a [Domain] (Subdomain, SecondLevelDomain,
//     TopLevelDomain) and [Domain.Apex]. weppos returns a DomainName with TLD,
//     SLD, TRD, and the matched Rule, plus Domain for the registrable name.
//   - This package always decomposes: a host that is itself a public suffix lands
//     wholly in TopLevelDomain, a single label lands in SecondLevelDomain, and an
//     IP literal yields a nil Domain. weppos reports an error for a name with no
//     registrable domain, such as a public suffix or a single label.
//   - Both fall back to the implicit "*" rule for unknown TLDs; weppos exposes
//     FindOptions.DefaultRule to change that behavior.
//   - Private suffixes are included by default. [WithIgnorePrivate] excludes them
//     here, so "google.blogspot.com" decomposes under "com"; weppos uses
//     FindOptions{IgnorePrivate: true} and Domain returns "blogspot.com".
//   - This package ships fixed embedded lists. weppos can load lists from files or
//     strings, keep multiple lists, and exposes AddRule and Size.
//   - This package embeds U-label and A-label rules and matches either form;
//     weppos stores A-labels and expects ASCII input.
//   - [Parser.Parse] handles ports, paths, relative references, and IP literals.
//     weppos treats its argument as a domain and also implements
//     net/http/cookiejar.PublicSuffixList, which this package does not.
//
// The runtime implementation relies on the Go standard library only. The
// internal/psl data generator is a separate nested module (cmd/pslgen) that
// uses golang.org/x/net/idna to produce A-labels, so it does not add a runtime
// dependency to the library.
//
// # Configuration
//
// Behavior is set through functional options passed to [New]:
//
//   - [WithDefaultScheme] applies a scheme to URLs that lack one, so "example.com/path"
//     parses as a host plus path instead of a relative reference. A scheme-less
//     "host:port" input (digits after the colon) is treated as a host and gets the
//     default scheme; any other URL whose first path segment contains ":" — an opaque
//     URI such as "mailto:user@example.com" — already carries a scheme per RFC 3986 and
//     is left unchanged.
//   - [WithIgnorePrivate] restricts the embedded Public Suffix List to its ICANN
//     section, excluding private-domain suffixes such as "blogspot.com". The implicit
//     "*" rule is preserved.
//   - [WithLenient] repairs invalid percent escapes before parsing, so
//     "https://example.com/%invalid" parses instead of failing. It does not relax host
//     validation, so an invalid authority such as "http://exa mple.com" still fails.
//
// A nil OptionFunc is ignored. The With* helpers cover common settings, and callers may
// write custom options because [Parser] is exported.
//
// # Concurrency
//
// A Parser is safe for concurrent use once configured: [Parser.Parse] only reads the
// configured settings. The Set* methods and options mutate those settings and must not
// run concurrently with parsing or with each other. Parsed [URL] and [Domain] values are
// independent and safe to share across goroutines.
package parser
