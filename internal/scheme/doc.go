// Package scheme holds the generated URL scheme tables used by the extractor.
// The tables are internal so the module's public API stays limited to the
// parser, extractor, and Public Suffix List packages.
//
// # Entry points
//
//   - [Official] is the generated table of active IANA-assigned schemes,
//     refreshed from the IANA registry by cmd/schemesgen.
//   - [Unofficial] is a hand-maintained table of widely used schemes that are
//     not IANA-registered (application- and service-specific protocols).
//   - [NoAuthority] is a hand-maintained table of schemes that address a
//     resource directly with a colon (e.g. "mailto:") rather than an
//     authority component.
package scheme
