// Package unicodes provides predefined sets of allowed Unicode character ranges for
// building regular expressions that match internationalized URLs.
//
// It suits work that must match IRIs (RFC 3987) rather than plain ASCII URIs, where a
// generated, verified set of character ranges is more useful than a hand-curated one.
//
// # Sets
//
//   - [AllowedUcsChar] holds the RFC 3987 ucschar ranges with separators excluded.
//   - [AllowedUcsCharMinusPunc] additionally excludes punctuation, for positions where a
//     trailing punctuation mark would otherwise be absorbed into a URL.
//
// Both constants are character-class contents: splice them into a "[...]" class, they
// are not standalone patterns. The sets are generated from the Unicode Character
// Database; see the module root for the go:generate directives.
//
// The implementation relies on the Go standard library only.
package unicodes
