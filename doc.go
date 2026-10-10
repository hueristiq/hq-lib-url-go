// Package url is the root of the hq-go-url module, a toolkit for working with URLs:
// extracting them from text and parsing them into granular components.
//
// It suits text processing, data mining, and reconnaissance work — anywhere URLs must
// be found in unstructured text or broken down into their constituent parts.
//
// # Packages
//
// The module is organized into focused packages:
//
//   - [parser]: parses raw URL strings into a [net/url]-compatible struct with the host
//     decomposed into subdomain, second-level domain, and top-level domain.
//   - [extractor]: compiles configurable regular expressions that extract URLs — with
//     or without schemes and hosts — and email addresses from text.
//   - [schemes]: categorized lists of URL schemes (official, unofficial, no-authority).
//   - [tlds]: categorized lists of top-level domains and public suffixes.
//   - [unicodes]: generated Unicode character ranges for matching internationalized
//     URLs.
//
// The data packages are regenerated from their public sources with go generate; see
// generate.go for the directives.
//
// [parser]: github.com/hueristiq/hq-go-url/parser
// [extractor]: github.com/hueristiq/hq-go-url/extractor
// [schemes]: github.com/hueristiq/hq-go-url/schemes
// [tlds]: github.com/hueristiq/hq-go-url/tlds
// [unicodes]: github.com/hueristiq/hq-go-url/unicodes
//
// The implementation relies on the Go standard library only.
package url
