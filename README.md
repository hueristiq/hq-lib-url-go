# hq-lib-url-go

![made with go](https://img.shields.io/badge/made%20with-Go-1E90FF.svg) [![go reference](https://pkg.go.dev/badge/github.com/hueristiq/hq-lib-url-go.svg)](https://pkg.go.dev/github.com/hueristiq/hq-lib-url-go) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=1E90FF)](https://github.com/hueristiq/hq-lib-url-go/blob/main/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-1E90FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/hq-lib-url-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-url-go/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/hq-lib-url-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-url-go/issues?q=is:issue+is:closed)

`hq-lib-url-go` is a [Go](https://golang.org/) package for working with URLs.

## Resources

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
	- [Quick start](#quick-start)
	- [Extraction](#extraction)
	- [Parsing](#parsing)
	- [Normalization](#normalization)
- [Contributing](#contributing)
- [Licensing](#licensing)

## Features

- **URL extraction:** longest-match patterns find fully qualified URLs, scheme-less hosts, email addresses, and relative paths, with functional options and exported pattern blocks.
- **PSL-aware parsing:** `parser` decomposes hosts into subdomain, SLD, and TLD against the embedded Public Suffix List, including multi-label and private suffixes, with label-aware matching.
- **Normalization:** lowercase scheme and host, default-port removal, dot-segment removal, and sorted query parameters through `URL.Normalize`.
- **Scheme inference:** `InferScheme` chooses `http` for local targets and `https` for everything else.
- **Typed URL values:** `parser.URL` embeds `net/url.URL` and exposes the decomposed `Domain`, while `Apex` returns the registrable domain.

## Installation

To install `hq-lib-url-go`, run the following command in your Go project:

```bash
go get github.com/hueristiq/hq-lib-url-go@latest
```

Requires Go 1.26.0 or later.

## Usage

The examples import the packages under the `hqgourl`-prefixed aliases.

### Quick start

The root package covers the common cases with the default parser and extractor.

```go
package main

import (
	"fmt"
	"log"

	hqgourl "github.com/hueristiq/hq-lib-url-go"
)

func main() {
	parsed, err := hqgourl.Parse("https://example.com/path?q=1")
	if err != nil {
		log.Fatal(err)
	}

	matches, err := hqgourl.Extract("see https://example.com/a and https://example.org/b")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(parsed.String())
	fmt.Println(matches)
}
```

### Extraction

The `extractor` package scans text and pulls out URLs with a composed regular expression. `Extract` is the one-call entry point. The pattern is compiled once per extractor and cached.

```go
package main

import (
	"fmt"
	"log"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
)

func main() {
	e := hqgourlextractor.New(hqgourlextractor.WithScheme())

	urls, err := e.Extract("Check these out: ftp://ftp.example.com, https://secure.example.com, and mailto:someone@example.com.")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(urls)
}
```

Configure the extractor with functional options. `WithScheme()` requires a scheme, and `WithSchemePattern` restricts schemes further:

```go
e := hqgourlextractor.New(
	hqgourlextractor.WithSchemePattern(`(?:https?|ftp)://`),
)
```

With no options, relative paths such as `/docs/guide` match too. `CompileRegex` returns the raw compiled pattern for callers who want to drive matching themselves.

The pattern building blocks are exported (`SchemePattern`, `KnownSchemePattern`, `IPv4Pattern`, `IPv6Pattern`, `PortPattern`, and others) for reuse in custom configurations:

```go
e := hqgourlextractor.New(
	hqgourlextractor.WithSchemePattern(hqgourlextractor.KnownSchemePattern),
)
```

### Parsing

The `parser` package extends Go's `net/url` parsing with a detailed domain breakdown.

```go
package main

import (
	"fmt"
	"log"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func main() {
	p := hqgourlparser.New(hqgourlparser.WithDefaultScheme("https"))

	parsed, err := p.Parse("www.example.com/path/file.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Scheme: %s\n", parsed.Scheme)
	fmt.Printf("Host: %s\n", parsed.Host)
	fmt.Printf("Hostname: %s\n", parsed.Hostname())
	fmt.Printf("Subdomain: %s\n", parsed.Domain.Subdomain)
	fmt.Printf("Second-level domain (SLD): %s\n", parsed.Domain.SecondLevelDomain)
	fmt.Printf("Top-level domain (TLD): %s\n", parsed.Domain.TopLevelDomain)
	fmt.Printf("Path: %s\n", parsed.Path)
}
```

Hosts are decomposed with the embedded Public Suffix List. IP literals get a nil `Domain`. A single-label host such as `localhost` lands wholly in `SecondLevelDomain`. The most specific matching rule, including private, wildcard, exception, and punycode rules, supplies `TopLevelDomain`, with the label before it in `SecondLevelDomain` and the rest in `Subdomain`. A host that is itself a public suffix, such as `co.uk`, lands wholly in `TopLevelDomain`. Unknown TLDs use the implicit `*` rule, so the last label becomes `TopLevelDomain`.

`Domain.Apex()` returns the registrable domain (SLD plus TLD). For `google.blogspot.com`, both this package and [`weppos/publicsuffix-go`](https://github.com/weppos/publicsuffix-go)'s `Domain` include the private `blogspot.com` suffix by default: `Apex()` is `google.blogspot.com`, and weppos returns `google.blogspot.com`. With `WithIgnorePrivate()`, this package gives `SLD="blogspot"`, `TLD="com"`, and `Apex() == "blogspot.com"`, while weppos with `IgnorePrivate: true` returns `blogspot.com`.

### Normalization

`URL.Normalize` rewrites a parsed URL in place into a canonical form. `InferScheme` guesses a scheme for a host that is missing one: `http` for local targets, `https` otherwise.

```go
package main

import (
	"fmt"
	"log"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func main() {
	p := hqgourlparser.New()

	parsed, err := p.Parse("https://WWW.EXAMPLE.COM:443/./a/../b/?b=2&a=1")
	if err != nil {
		log.Fatal(err)
	}

	parsed.Normalize()

	fmt.Println(parsed.String())
	fmt.Println(hqgourlparser.InferScheme("localhost:8080"))
	fmt.Println(hqgourlparser.InferScheme("example.com"))
}
```

## Contributing

Contributions are welcome and encouraged! Feel free to submit [Pull Requests](https://github.com/hueristiq/hq-lib-url-go/pulls) or report [Issues](https://github.com/hueristiq/hq-lib-url-go/issues).

A big thank you to all the [contributors](https://github.com/hueristiq/hq-lib-url-go/graphs/contributors) for your ongoing support!

![contributors](https://contrib.rocks/image?repo=hueristiq/hq-lib-url-go&max=500)

## Licensing

This package is licensed under the [MIT license](https://opensource.org/license/mit). You are free to use, modify, and distribute it, as long as you follow the terms of the license. You can find the full license text in the repository - [Full MIT license text](https://github.com/hueristiq/hq-lib-url-go/blob/main/LICENSE).
