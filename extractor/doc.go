// Package extractor extracts URLs from plain text using regular expressions.
//
// It suits scanning unstructured text — web pages, logs, documents — for URLs, where a
// single composed pattern that understands schemes, hosts, ports, Unicode domains, and
// email addresses is more useful than hand-rolled splitting.
//
// # Entry points
//
// [New] returns an [Extractor], and [Extractor.CompileRegex] compiles the configured
// extractor into a [regexp.Regexp] tuned to find the longest match at each position,
// returning an error when a custom pattern fails to compile.
//
// # Extraction
//
// The composed pattern understands schemes, hosts, ports, Unicode domains, and email
// addresses. With no options, relative paths (e.g. "/docs/guide") match too.
//
// # Patterns
//
// The building blocks of the composed pattern are exported for reuse in custom
// configurations: [SchemePattern], [KnownOfficialSchemePattern],
// [KnownUnofficialSchemePattern], [KnownNoAuthoritySchemePattern], [KnownSchemePattern],
// [IPv4Pattern], [NonEmptyIPv6Pattern], [IPv6Pattern], [PortPattern], and
// [PortOptionalPattern]. They are variables only so they can be spliced into larger
// pattern strings; callers must not modify them.
//
// Host matching is anchored on known TLDs: the official and pseudo TLD lists from the
// tlds package are split once at startup by whether an entry ends in an ASCII word
// character — only those entries are followed by a "\b" word boundary, since Go's "\b"
// does not recognize non-ASCII word characters — and the resulting alternation is
// composed once, on first use, so the compiled pattern never depends on the lists' sort
// order and compilation does not rebuild it per call. Punycode (xn--) hosts are matched
// even though the generated official TLD list excludes them.
//
// # Unescaping
//
// [DecodeJSEscapes] converts the escape sequences found in JavaScript and JSON strings
// into the characters they represent, so that matching the decoded text recovers
// escaped URLs (for example `https:\/\/example.com`, `\x68ttps://example.com`, or
// `example\u002ecom`).
//
// # Literals
//
// [JSLiteralGroups] finds string literals in JavaScript source text and collapses runs
// of concatenated literals into single strings. Code like
//
//	const u = "https://" + host + "/api";
//
// yields a [LiteralGroup] whose [LiteralGroup.Collapsed] value is "https://EXPR/api",
// with [ExpressionPlaceholder] standing in for the expression whose value cannot be
// known. Collapsed strings are then matchable by URL patterns, recovering URLs that a
// plain scan of the raw text would miss. The scanner is text-based rather than a
// parser: division expressions, regex literals, and unquoted markup may occasionally
// be read as literals.
//
// Literals inside an unknown expression are returned as separate groups, which may
// overlap the enclosing group's Start/End range.
//
// The implementation relies on the Go standard library only.
//
// # Configuration
//
// Behavior is set through functional options passed to [New]:
//
//   - [WithScheme] requires a scheme, so only fully-qualified URLs match.
//   - [WithSchemePattern] restricts the scheme with a custom pattern, e.g.
//     "(?:https?://)".
//   - [WithHost] additionally matches scheme-less URLs with a host and bare email
//     addresses; emails land in the "relaxedEmail" subexpression.
//   - [WithHostPattern] restricts the host with a custom pattern.
//   - [WithoutEmail] excludes bare email addresses.
//   - [WithoutUserInfo] excludes user info in the authority.
//
// A nil OptionFunc is ignored. The With* and Without* helpers cover common settings, and
// callers may write custom options because [Extractor] is exported.
//
// # Concurrency
//
// An Extractor is safe for concurrent use once configured: [Extractor.CompileRegex] only
// reads the configured settings. The Set* methods and options mutate those settings and
// must not run concurrently with compilation or with each other.
package extractor
