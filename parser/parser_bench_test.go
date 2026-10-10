package parser_test

import (
	"testing"

	"github.com/hueristiq/hq-go-url/parser"
)

// parseURLSink and errParseSink keep benchmark results alive so the compiler cannot
// optimize the measured calls away.
var (
	parseURLSink *parser.URL
	errParseSink error
)

// Benchmark_Parser_Parse measures the full parse path — scheme application, net/url
// parsing, and TLD decomposition — on a URL exercising every component: user info, a
// subdomain, a multi-label TLD, a port, a query, and a fragment.
func Benchmark_Parser_Parse(b *testing.B) {
	p := parser.New()

	b.ReportAllocs()

	var (
		parsed *parser.URL
		err    error
	)

	for b.Loop() {
		parsed, err = p.Parse("https://user:pass@www.example.co.uk:8443/path/to/resource?a=1&b=2#section")
	}

	parseURLSink = parsed
	errParseSink = err
}

// Benchmark_Parser_ParseSingleLabelHost measures the parse path for a single-label host
// ("localhost"), which skips TLD lookup entirely.
func Benchmark_Parser_ParseSingleLabelHost(b *testing.B) {
	p := parser.New()

	b.ReportAllocs()

	var (
		parsed *parser.URL
		err    error
	)

	for b.Loop() {
		parsed, err = p.Parse("http://localhost:8080/health")
	}

	parseURLSink = parsed
	errParseSink = err
}
