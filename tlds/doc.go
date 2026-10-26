// Package tlds provides categorized lists of top-level domains (TLDs) and effective
// TLDs (eTLDs): official IANA-approved domains and public suffixes on one hand, widely
// used pseudo-TLDs on the other.
//
// It suits work that must decompose a hostname into its registrable parts — subdomain,
// second-level domain, and suffix — where a maintained list of known suffixes is more
// useful than positional guessing.
//
// # Entry points
//
// [Official] and [Pseudo] are the exported TLD lists. They are read-only shared data;
// do not modify them.
//
// # Lists
//
//   - [Official] holds the union of the IANA TLD list and the ICANN section of the
//     Public Suffix List, so multi-label suffixes such as "co.uk" are recognized
//     alongside single-label TLDs such as "com". Punycode (xn--) TLDs are excluded from
//     the generated list; the extractor package matches them anyway.
//   - [Pseudo] holds unofficial but widely recognized suffixes used in private networks
//     and alternative name systems, such as "local" and "localhost".
//
// Both lists are sorted. The official list is generated from public data sources; see
// the module root for the go:generate directives.
//
// The implementation relies on the Go standard library only.
package tlds
