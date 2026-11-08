// Package tld holds the generated top-level-domain tables used by the
// extractor. The tables are internal so the module's public API stays limited
// to the parser, extractor, and Public Suffix List packages.
//
// # Entry points
//
//   - [Official] is the generated table of public top-level domains and
//     effective TLDs, refreshed from the IANA and Public Suffix List sources
//     by cmd/tldsgen.
//   - [Pseudo] is a hand-maintained table of widely used unofficial or
//     pseudo-TLDs (special-use and decentralized names).
package tld
