package main

import (
	"fmt"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func main() {
	parsed, err := hqgourlparser.New().Parse("https://sub.example.co.uk/path")
	if err != nil {
		fmt.Println("parse:", err)

		return
	}

	fmt.Println(parsed.Host, parsed.Domain.SecondLevelDomain, parsed.Domain.TopLevelDomain)

	matches, err := hqgourlextractor.New().Extract("visit https://example.com and http://test.org")
	if err != nil {
		fmt.Println("extract:", err)

		return
	}

	fmt.Println(matches)
}
