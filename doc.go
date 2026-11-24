// Package url is the root of the hq-lib-url-go module, a toolkit for working with URLs:
// extracting them from text and parsing them into granular components.
//
// It suits text processing, data mining, and reconnaissance work — anywhere URLs must
// be found in unstructured text or broken down into their constituent parts.
//
// # Packages
//
// The module is organized into focused packages, which carry the entire API:
//
//   - [parser]: parses raw URL strings into a [net/url]-compatible struct with the host
//     decomposed into subdomain, second-level domain, and top-level domain.
//     [parser.New] and [parser.Parser.Parse] support custom options such as
//     a default scheme for scheme-less input.
//   - [extractor]: compiles configurable regular expressions that extract URLs — with
//     or without schemes and hosts — and email addresses from text.
//
// The PSL lookup lives in internal/psl, and the generated scheme,
// top-level-domain, and Unicode tables live in internal/scheme, internal/tld,
// and internal/unicode. They are regenerated with go generate; see generate.go
// for the directives.
//
// [parser]: github.com/hueristiq/hq-lib-url-go/parser
// [extractor]: github.com/hueristiq/hq-lib-url-go/extractor
// [net/url]: net/url
package url
