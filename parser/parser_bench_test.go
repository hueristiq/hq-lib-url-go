package parser_test

import (
	"testing"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

var (
	parseURLSink *hqgourlparser.URL
	errParseSink error
)

func Benchmark_Parser_Parse(b *testing.B) {
	p := hqgourlparser.New()

	b.ReportAllocs()

	var (
		parsed *hqgourlparser.URL
		err    error
	)

	for range b.N {
		parsed, err = p.Parse("https://user:pass@www.example.co.uk:8443/path/to/resource?a=1&b=2#section")
	}

	parseURLSink = parsed
	errParseSink = err
}

func Benchmark_Parser_ParseSingleLabelHost(b *testing.B) {
	p := hqgourlparser.New()

	b.ReportAllocs()

	var (
		parsed *hqgourlparser.URL
		err    error
	)

	for range b.N {
		parsed, err = p.Parse("http://localhost:8080/health")
	}

	parseURLSink = parsed
	errParseSink = err
}
