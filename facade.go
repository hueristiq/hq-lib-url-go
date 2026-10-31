package url

import (
	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

// Parse decomposes raw into a structured URL with the default parser:
// PSL-aware domain decomposition. Use parser.New with
// parser.WithDefaultScheme when scheme-less input must become absolute.
//
// Parameters:
//   - raw (string): the raw URL string to parse.
//
// Returns:
//   - parsed (*hqgourlparser.URL): the structured URL with the embedded
//     net/url.URL and domain details.
//   - err (error): non-nil when raw cannot be parsed.
func Parse(raw string) (parsed *hqgourlparser.URL, err error) {
	return hqgourlparser.New().Parse(raw)
}

// Extract finds every URL in text with the default extractor.
//
// Parameters:
//   - text (string): the text to search.
//
// Returns:
//   - matches ([]string): the URLs found, in order of appearance; empty when
//     none match.
//   - err (error): non-nil when the composed pattern fails to compile.
func Extract(text string) (matches []string, err error) {
	return hqgourlextractor.New().Extract(text)
}
