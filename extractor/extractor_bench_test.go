package extractor_test

import (
	"testing"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
)

var (
	extractorSink *hqgourlextractor.Extractor
	errSink       error
	matchesSink   []string
)

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()

	var extractor *hqgourlextractor.Extractor

	for range b.N {
		extractor, errSink = hqgourlextractor.New()
	}

	extractorSink = extractor
}

func BenchmarkNewWithScheme(b *testing.B) {
	b.ReportAllocs()

	var extractor *hqgourlextractor.Extractor

	for range b.N {
		extractor, errSink = hqgourlextractor.New(hqgourlextractor.WithScheme())
	}

	extractorSink = extractor
}

func BenchmarkExtractWithHost(b *testing.B) {
	extractor := mustNew(b, hqgourlextractor.WithHost())

	b.ReportAllocs()

	var matches []string

	for range b.N {
		matches = extractor.Extract(corpus)
	}

	matchesSink = matches
}

func BenchmarkExtract(b *testing.B) {
	extractor := mustNew(b)

	b.ReportAllocs()

	var matches []string

	for range b.N {
		matches = extractor.Extract(corpus)
	}

	matchesSink = matches
}
