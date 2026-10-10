// Package extractor extracts URLs from plain text using regular expressions.
//
// It suits scanning unstructured text — web pages, logs, documents — for URLs, where a
// single composed pattern that understands schemes, hosts, ports, Unicode domains, and
// email addresses is more useful than hand-rolled splitting.
//
// # Extraction
//
// [New] returns an [Extractor]; [Extractor.CompileRegex] compiles the configured
// extractor into a [regexp.Regexp] tuned to find the longest match at each position.
// The behavior is set through functional options:
//
//   - [WithScheme] requires a scheme, so only fully-qualified URLs match.
//   - [WithSchemePattern] restricts the scheme with a custom pattern, e.g. "(?:https?://)".
//   - [WithHost] additionally matches scheme-less URLs with a host and bare email
//     addresses; emails land in the "relaxedEmail" subexpression.
//   - [WithHostPattern] restricts the host with a custom pattern.
//   - With no options, relative paths (e.g. "/docs/guide") match too.
//
// # Patterns
//
// The building blocks of the composed pattern are exported for reuse in custom
// configurations: [ExtractorSchemePattern], [ExtractorKnownOfficialSchemePattern],
// [ExtractorKnownUnofficialSchemePattern], [ExtractorKnownNoAuthoritySchemePattern],
// [ExtractorKnownSchemePattern], [ExtractorIPv4Pattern], [ExtractorNonEmptyIPv6Pattern],
// [ExtractorIPv6Pattern], [ExtractorPortPattern], and [ExtractorPortOptionalPattern].
//
// Host matching is anchored on known TLDs: the official and pseudo TLD lists from the
// tlds package are split once at startup by whether an entry ends in an ASCII word
// character — only those entries are followed by a "\b" word boundary, since Go's "\b"
// does not recognize non-ASCII word characters — so the compiled pattern never depends
// on the lists' sort order.
//
// The implementation relies on the Go standard library only.
package extractor
