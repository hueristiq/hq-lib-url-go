// Package url is the root of the hq-lib-url-go module, a toolkit for working with URLs:
// extracting them from text and parsing them into granular components.
//
// It suits text processing, data mining, and reconnaissance work — anywhere URLs must
// be found in unstructured text or broken down into their constituent parts.
//
// # Entry points
//
// [Parse] and [Extract] cover the common cases with the default parser and
// extractor. Use [parser.New] and [parser.Parser.Parse] for structured
// decomposition with custom options, or [extractor.New] and
// [extractor.Extractor.CompileRegex] to build the regular expression that
// finds URLs in text.
//
// # Packages
//
// The module is organized into focused packages:
//
//   - [parser]: parses raw URL strings into a [net/url]-compatible struct with the host
//     decomposed into subdomain, second-level domain, and top-level domain.
//   - [extractor]: compiles configurable regular expressions that extract URLs — with
//     or without schemes and hosts — and email addresses from text.
//
// The PSL lookup lives in internal/psl, and the generated scheme,
// top-level-domain, and Unicode tables live in [scheme], [tld], and [unicode].
// They are regenerated with go generate; see generate.go for the directives. The
// PSL generator is a separate nested module so the library module itself has no
// runtime dependencies beyond the standard library.
//
// [parser]: github.com/hueristiq/hq-lib-url-go/parser
// [extractor]: github.com/hueristiq/hq-lib-url-go/extractor
// [scheme]: github.com/hueristiq/hq-lib-url-go/scheme
// [tld]: github.com/hueristiq/hq-lib-url-go/tld
// [unicode]: github.com/hueristiq/hq-lib-url-go/unicode
package url
