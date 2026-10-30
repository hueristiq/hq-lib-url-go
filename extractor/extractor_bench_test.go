package extractor_test

import (
	"regexp"
	"testing"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
)

// regexSink, errRegexSink, and matchesSink keep benchmark results alive so the compiler
// cannot optimize the measured calls away.
var (
	regexSink    *regexp.Regexp
	errRegexSink error
	matchesSink  []string
)

// Benchmark_Extractor_CompileRegex measures the cost of composing and compiling the
// default (scheme-less-tolerant) extraction regex, which splices the full known-TLD
// alternation into the pattern.
func Benchmark_Extractor_CompileRegex(b *testing.B) {
	e := hqgourlextractor.New()

	b.ReportAllocs()

	var (
		regex *regexp.Regexp
		err   error
	)

	for range b.N {
		regex, err = e.CompileRegex()
	}

	regexSink = regex
	errRegexSink = err
}

// Benchmark_Extractor_CompileRegexWithScheme measures the cost of composing and
// compiling the strictest (scheme-required) extraction regex.
func Benchmark_Extractor_CompileRegexWithScheme(b *testing.B) {
	e := hqgourlextractor.New(hqgourlextractor.WithScheme())

	b.ReportAllocs()

	var (
		regex *regexp.Regexp
		err   error
	)

	for range b.N {
		regex, err = e.CompileRegex()
	}

	regexSink = regex
	errRegexSink = err
}

// Benchmark_Extractor_FindAllString measures the match hot path: scanning a realistic
// mixed corpus with the host-requiring regex.
func Benchmark_Extractor_FindAllString(b *testing.B) {
	regex, err := hqgourlextractor.New(hqgourlextractor.WithHost()).CompileRegex()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	var matches []string

	for range b.N {
		matches = regex.FindAllString(corpus, -1)
	}

	matchesSink = matches
}

// BenchmarkExtract measures the cached one-call extraction path: the pattern
// is compiled once and every call scans the corpus.
func BenchmarkExtract(b *testing.B) {
	e := hqgourlextractor.New()

	b.ReportAllocs()

	var matches []string

	for range b.N {
		matches, _ = e.Extract(corpus)
	}

	matchesSink = matches
}
