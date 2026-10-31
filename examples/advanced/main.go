package main

import (
	"fmt"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func main() {
	parser := hqgourlparser.New(
		hqgourlparser.WithDefaultScheme("https"),
		hqgourlparser.WithIgnorePrivate(),
	)

	parsed, err := parser.Parse("example.com/path")
	if err != nil {
		fmt.Println("parse:", err)

		return
	}

	parsed.Normalize()
	fmt.Println(parsed.Scheme, parsed.Host)

	extractor := hqgourlextractor.New(
		hqgourlextractor.WithScheme(),
		hqgourlextractor.WithoutEmail(),
	)

	matches, err := extractor.Extract("mailto:user@example.com https://example.com/path")
	if err != nil {
		fmt.Println("extract:", err)

		return
	}

	fmt.Println(matches)
}
