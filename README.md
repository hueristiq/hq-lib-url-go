# hq-lib-url-go

![made with go](https://img.shields.io/badge/made%20with-Go-1E90FF.svg) [![go reference](https://pkg.go.dev/badge/github.com/hueristiq/hq-lib-url-go.svg)](https://pkg.go.dev/github.com/hueristiq/hq-lib-url-go) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=1E90FF)](https://github.com/hueristiq/hq-lib-url-go/blob/main/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-1E90FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/hq-lib-url-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-url-go/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/hq-lib-url-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-url-go/issues?q=is:issue+is:closed) [![contribution](https://img.shields.io/badge/contributions-welcome-1E90FF.svg)](https://github.com/hueristiq/hq-lib-url-go/blob/main/CONTRIBUTING.md)

`hq-lib-url-go` is a [Go](https://golang.org/) module for working with URLs: it extracts them from plain text with configurable regular expressions, and it parses them into granular components — including a domain broken down into subdomain, second-level domain, and top-level domain.

## Resources

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
	- [Extraction](#extraction)
	- [Parsing](#parsing)
	- [Key Access](#key-access)
	- [Normalization](#normalization)
- [Data Packages](#data-packages)
- [Contributing](#contributing)
- [Licensing](#licensing)

## Features

- **URL Extraction:** Composed, longest-match regular expressions that find fully-qualified URLs, scheme-less host URLs, email addresses, and relative paths in any text, plus fine-grained functional options and exported pattern building blocks.
- **URL Parsing:** Extends [`net/url`](https://pkg.go.dev/net/url) with a decomposition of the host into subdomain, SLD, and TLD, matched exactly against official and pseudo TLD lists — including multi-label public suffixes such as `co.uk` — via binary search.
- **Key Access:** Component lookup by name (`scheme`, `host`, `port`, `subdomain`, `sld`, `tld`, `apex`, `path`, `query`, `fragment`, and more).
- **Normalization:** Canonicalization — lowercase scheme and host, default-port removal, dot-segment removal, sorted query parameters — plus per-host scheme inference with `InferScheme`.
- **Standard Library Only:** No third-party runtime dependencies; the `schemes`, `tlds`, and `unicodes` data packages are regenerated from public sources with `go generate`.

## Installation

To install `hq-lib-url-go`, run the following command in your Go project:

```bash
go get -v -u github.com/hueristiq/hq-lib-url-go
```

## Usage

The examples below import the packages under `hqgourl`-prefixed aliases:

```go
import (
	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)
```

### Extraction

The `extractor` package scans text and pulls out URLs using a composed regular expression. Configure it with functional options — here `WithScheme()`, which requires a scheme:

```go
package main

import (
	"fmt"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
)

func main() {
	e := hqgourlextractor.New(hqgourlextractor.WithScheme())

	regex := e.CompileRegex()

	text := "Check these out: ftp://ftp.example.com, https://secure.example.com, and mailto:someone@example.com."

	urls := regex.FindAllString(text, -1)

	fmt.Println("Extracted URLs:")

	for _, u := range urls {
		fmt.Println(u)
	}
}
```

Available configurations:

- `WithScheme()` — require a scheme, so only fully-qualified URLs match.
- `WithHost()` — additionally match scheme-less host URLs and bare email addresses (emails land in the `relaxedEmail` subexpression).
- `WithSchemePattern(pattern)` / `WithHostPattern(pattern)` — restrict the scheme or host with a custom pattern, e.g.:

	```go
	e := hqgourlextractor.New(
		hqgourlextractor.WithSchemePattern(`(?:https?|ftp)://`),
	)
	```

- With no options, relative paths such as `/docs/guide` match too.

The pattern building blocks are exported (`ExtractorSchemePattern`, `ExtractorKnownSchemePattern`, `ExtractorIPv4Pattern`, `ExtractorIPv6Pattern`, `ExtractorPortPattern`, …) for reuse in custom configurations:

```go
e := hqgourlextractor.New(
	hqgourlextractor.WithSchemePattern(hqgourlextractor.ExtractorKnownSchemePattern),
)
```

### Parsing

The `parser` package extends Go's `net/url` parsing with a detailed domain breakdown:

```go
package main

import (
	"fmt"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func main() {
	p := hqgourlparser.New(hqgourlparser.WithDefaultScheme("https"))

	parsed, err := p.Parse("www.example.com:8080/path/file.txt")
	if err != nil {
		fmt.Println("Error parsing URL:", err)

		return
	}

	fmt.Printf("Scheme: %s\n", parsed.Scheme)
	fmt.Printf("Host: %s\n", parsed.Host)
	fmt.Printf("Hostname: %s\n", parsed.Hostname())
	fmt.Printf("Subdomain: %s\n", parsed.Domain.Subdomain)
	fmt.Printf("Second-Level Domain (SLD): %s\n", parsed.Domain.SecondLevelDomain)
	fmt.Printf("Top-Level Domain (TLD): %s\n", parsed.Domain.TopLevelDomain)
	fmt.Printf("Port: %s\n", parsed.Port())
	fmt.Printf("Path: %s\n", parsed.Path)
}
```

Available configurations:

- `WithDefaultScheme(scheme)` — parse scheme-less URLs like `example.com/path` as absolute URLs with the given scheme.
- `WithTLDs("custom", ...)` — replace the recognized TLD set (the union of the official and pseudo lists by default).

Hosts are decomposed deterministically: IP literals get a nil `Domain`; a single-label host such as `localhost` lands wholly in `SecondLevelDomain`; a host with no recognized TLD is decomposed positionally with `TopLevelDomain` left empty; a host that is entirely a known suffix (e.g. `co.uk`) lands wholly in `TopLevelDomain`.

### Key Access

`URL.Key` returns a single component by name:

```go
package main

import (
	"fmt"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func main() {
	p := hqgourlparser.New()

	parsed, err := p.Parse("https://user:pass@www.example.co.uk:8443/users?id=123#about")
	if err != nil {
		fmt.Println("Error parsing URL:", err)

		return
	}

	fmt.Println(parsed.Key("scheme"))    // "https"
	fmt.Println(parsed.Key("host"))      // "www.example.co.uk:8443"
	fmt.Println(parsed.Key("subdomain")) // "www"
	fmt.Println(parsed.Key("sld"))       // "example" ("root" is an alias)
	fmt.Println(parsed.Key("tld"))       // "co.uk"
	fmt.Println(parsed.Key("apex"))      // "example.co.uk"
	fmt.Println(parsed.Key("path"))      // "/users"
	fmt.Println(parsed.Key("query"))     // "id=123"
}
```

Supported keys: `scheme`, `user`, `pass`, `host`, `port`, `domain`, `subdomain`, `sld` (alias `root`), `tld`, `apex`, `path`, `query`, `fragment`. An unrecognized or absent component yields an empty string.

### Normalization

`URL.Normalize` rewrites a parsed URL in place into a canonical form, and `InferScheme` guesses a scheme for a host that is missing one (`http` for local targets, `https` otherwise):

```go
package main

import (
	"fmt"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func main() {
	p := hqgourlparser.New()

	parsed, err := p.Parse("https://WWW.EXAMPLE.COM:443/./a/../b/?b=2&a=1")
	if err != nil {
		fmt.Println("Error parsing URL:", err)

		return
	}

	parsed.Normalize()

	fmt.Println(parsed.String()) // "https://www.example.com/b/?a=1&b=2"

	fmt.Println(hqgourlparser.InferScheme("localhost:8080")) // "http"
	fmt.Println(hqgourlparser.InferScheme("example.com"))    // "https"
}
```

## Data Packages

The `schemes`, `tlds`, and `unicodes` packages hold the categorized data the parser and extractor build on: official, unofficial, and no-authority URL schemes; official and pseudo TLDs (including multi-label public suffixes); and RFC 3987 Unicode character ranges. They are regenerated from public sources with:

```bash
go generate ./...
```

## Contributing

Contributions are welcome and encouraged! Feel free to submit [Pull Requests](https://github.com/hueristiq/hq-lib-url-go/pulls) or report [Issues](https://github.com/hueristiq/hq-lib-url-go/issues). For more details, check out the [contribution guidelines](https://github.com/hueristiq/hq-lib-url-go/blob/main/CONTRIBUTING.md).

A big thank you to all the [contributors](https://github.com/hueristiq/hq-lib-url-go/graphs/contributors) for your ongoing support!

![contributors](https://contrib.rocks/image?repo=hueristiq/hq-lib-url-go&max=500)

## Licensing

This package is licensed under the [MIT license](https://opensource.org/license/mit). You are free to use, modify, and distribute it, as long as you follow the terms of the license. You can find the full license text in the repository - [Full MIT license text](https://github.com/hueristiq/hq-lib-url-go/blob/main/LICENSE).
