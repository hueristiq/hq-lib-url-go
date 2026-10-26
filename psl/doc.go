// Package psl implements the Public Suffix List matching algorithm used by
// [github.com/hueristiq/hq-lib-url-go/parser].
//
// # Entry points
//
// [Default] contains the full ICANN and private sections, while [ICANN] excludes
// private-domain suffixes. [List.Find] locates the most specific rule matching a
// hostname, and [Rule.Decompose] splits a host into the labels before the public suffix
// and the suffix itself.
//
// Rules come in three kinds: normal suffixes such as "co.uk", wildcard rules
// such as "*.ck", and exception rules such as "!www.ck". The most specific
// rule wins; when none matches, both lists fall back to the implicit "*" rule.
//
// # Concurrency
//
// The lists and their [Rule] values are immutable after construction and safe for
// concurrent use.
package psl
