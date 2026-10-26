// Package schemes provides categorized lists of URL schemes: official IANA-assigned
// schemes, commonly used unofficial schemes, and schemes that do not require an
// authority component.
//
// It suits work that must recognize or validate the scheme of a URL — crawlers,
// extractors, and link classifiers — where a maintained, categorized list is more
// useful than a hand-written one.
//
// # Entry points
//
// [Official], [Unofficial], and [NoAuthority] are the exported scheme lists. They are
// read-only shared data; do not modify them.
//
// # Lists
//
//   - [Official] holds the active IANA-assigned schemes — Historical and otherwise
//     obsolete registry entries are excluded — and is regenerated periodically from the
//     IANA registry.
//   - [Unofficial] holds widely used schemes that are not registered with IANA.
//   - [NoAuthority] holds schemes followed by ":" instead of "//", such as "mailto".
//
// All lists are sorted. The official list is generated from the IANA registry; see the
// module root for the go:generate directives.
//
// The implementation relies on the Go standard library only.
package schemes
