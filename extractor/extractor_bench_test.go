package extractor_test

import (
	"regexp"
	"testing"

	"github.com/hueristiq/hq-lib-url-go/extractor"
)

// regexSink and matchesSink keep benchmark results alive so the compiler cannot
// optimize the measured calls away.
var (
	regexSink   *regexp.Regexp
	matchesSink []string
)

// Benchmark_Extractor_CompileRegex measures the cost of composing and compiling the
// default (scheme-less-tolerant) extraction regex, which splices the full known-TLD
// alternation into the pattern.
func Benchmark_Extractor_CompileRegex(b *testing.B) {
	e := extractor.New()

	b.ReportAllocs()

	var regex *regexp.Regexp

	for b.Loop() {
		regex = e.CompileRegex()
	}

	regexSink = regex
}

// Benchmark_Extractor_CompileRegexWithScheme measures the cost of composing and
// compiling the strictest (scheme-required) extraction regex.
func Benchmark_Extractor_CompileRegexWithScheme(b *testing.B) {
	e := extractor.New(extractor.WithScheme())

	b.ReportAllocs()

	var regex *regexp.Regexp

	for b.Loop() {
		regex = e.CompileRegex()
	}

	regexSink = regex
}

// Benchmark_Extractor_FindAllString measures the match hot path: scanning a realistic
// mixed corpus with the host-requiring (Relaxed) regex.
func Benchmark_Extractor_FindAllString(b *testing.B) {
	regex := extractor.New(extractor.Relaxed()).CompileRegex()

	b.ReportAllocs()

	var matches []string

	for b.Loop() {
		matches = regex.FindAllString(corpus, -1)
	}

	matchesSink = matches
}
