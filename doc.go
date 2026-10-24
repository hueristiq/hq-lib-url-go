// Package url is the root of the hq-lib-url-go module, a toolkit for working with URLs:
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
//   - [psl]: matches hostnames against the Public Suffix List and holds the generated
//     rule table used by [parser].
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
// [parser]: github.com/hueristiq/hq-lib-url-go/parser
// [psl]: github.com/hueristiq/hq-lib-url-go/psl
// [extractor]: github.com/hueristiq/hq-lib-url-go/extractor
// [schemes]: github.com/hueristiq/hq-lib-url-go/schemes
// [tlds]: github.com/hueristiq/hq-lib-url-go/tlds
// [unicodes]: github.com/hueristiq/hq-lib-url-go/unicodes
//
// The library packages rely on the Go standard library only; the PSL generator uses
// golang.org/x/net/idna to produce A-labels.
package url
