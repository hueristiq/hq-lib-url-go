# hq-lib-url-go

![made with go](https://img.shields.io/badge/made%20with-Go-1E90FF.svg) [![go reference](https://pkg.go.dev/badge/github.com/hueristiq/hq-lib-url-go.svg)](https://pkg.go.dev/github.com/hueristiq/hq-lib-url-go) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=1E90FF)](https://github.com/hueristiq/hq-lib-url-go/blob/main/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-1E90FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/hq-lib-url-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-url-go/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/hq-lib-url-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-url-go/issues?q=is:issue+is:closed) [![contribution](https://img.shields.io/badge/contributions-welcome-1E90FF.svg)](https://github.com/hueristiq/hq-lib-url-go/blob/main/CONTRIBUTING.md)

`hq-lib-url-go` is a [Go](https://golang.org/) package for working with URLs: it extracts them from text with configurable regular expressions, and parses them into granular components.

## Resources

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
	- [Extraction](#extraction)
	- [Parsing](#parsing)
		- [Key Access](#key-access)
	- [Normalization](#normalization)
- [Contributing](#contributing)
- [Licensing](#licensing)

## Features

- **URL Extraction:** Composed, longest-match regular expressions that find fully-qualified URLs, scheme-less host URLs, email addresses, and relative paths in any text, plus fine-grained functional options and exported pattern building blocks.
- **URL Parsing:** Extends [`net/url`](https://pkg.go.dev/net/url) with a decomposition of the host into subdomain, SLD, and TLD, matched exactly against the embedded Public Suffix List — including multi-label public suffixes such as `co.uk` — with label-aware lookups that never match inside a longer suffix.
- **Normalization:** Canonicalization — lowercase scheme and host, default-port removal, dot-segment removal, sorted query parameters — plus per-host scheme inference with `InferScheme`.

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
	"log"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
)

func main() {
	e := hqgourlextractor.New(hqgourlextractor.WithScheme())

	regex, err := e.CompileRegex()
	if err != nil {
		log.Fatal(err)
	}

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

The pattern building blocks are exported (`SchemePattern`, `KnownSchemePattern`, `IPv4Pattern`, `IPv6Pattern`, `PortPattern`, …) for reuse in custom configurations:

```go
e := hqgourlextractor.New(
	hqgourlextractor.WithSchemePattern(hqgourlextractor.KnownSchemePattern),
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

	parsed, err := p.Parse("www.example.com/path/file.txt")
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
	fmt.Printf("Path: %s\n", parsed.Path)
}
```

Available configurations:

- `WithDefaultScheme(scheme)` — parse scheme-less URLs like `example.com/path` as absolute URLs with the given scheme. A scheme-less `host:port` input (digits after the colon, e.g. `example.com:8080/api`) is treated as a host and gets the default scheme; any other first-path-segment colon — an opaque URI such as `mailto:user@example.com` — already carries a scheme and is left unchanged.
- `WithIgnorePrivate()` — use only the ICANN section of the embedded Public Suffix List, excluding private suffixes such as `blogspot.com` and `s3.amazonaws.com`. The implicit `*` rule is kept, so unknown TLDs still become `TopLevelDomain`. For example, `foo.blogspot.com` yields `Subdomain="foo"`, `SLD="blogspot"`, `TLD="com"` instead of `Subdomain=""`, `SLD="foo"`, `TLD="blogspot.com"`.
- `WithLenient()` — when `net/url.Parse` rejects a URL because of an invalid percent escape such as `https://example.com/%invalid`, repair the escape with `RepairInvalidEscapes` and parse again. Strict parsing remains the default. Lenient mode does not relax host validation, so `http://exa mple.com` still fails.

Hosts are decomposed with the embedded Public Suffix List: IP literals get a nil `Domain`; a single-label host such as `localhost` lands wholly in `SecondLevelDomain`; the most specific matching rule — including private, wildcard, exception, and punycode rules — supplies `TopLevelDomain`, with the label before it in `SecondLevelDomain` and the rest in `Subdomain`; a host that is itself a public suffix (e.g. `co.uk`) lands wholly in `TopLevelDomain`. Unknown TLDs use the implicit `*` rule, so the last label becomes `TopLevelDomain`.

`Domain.Apex()` returns the registrable domain (SLD plus TLD). For `google.blogspot.com`, both this package and [`weppos/publicsuffix-go`](https://github.com/weppos/publicsuffix-go)'s `Domain` include the private `blogspot.com` suffix by default: `Apex()` is `google.blogspot.com`, and weppos returns `google.blogspot.com`. With `WithIgnorePrivate()`, this package gives `SLD="blogspot"`, `TLD="com"`, and `Apex() == "blogspot.com"`, while weppos with `IgnorePrivate: true` returns `blogspot.com`.

#### Key Access

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

## Contributing

Contributions are welcome and encouraged! Feel free to submit [Pull Requests](https://github.com/hueristiq/hq-lib-url-go/pulls) or report [Issues](https://github.com/hueristiq/hq-lib-url-go/issues). For more details, check out the [contribution guidelines](https://github.com/hueristiq/hq-lib-url-go/blob/main/CONTRIBUTING.md).

A big thank you to all the [contributors](https://github.com/hueristiq/hq-lib-url-go/graphs/contributors) for your ongoing support!

![contributors](https://contrib.rocks/image?repo=hueristiq/hq-lib-url-go&max=500)

## Licensing

This package is licensed under the [MIT license](https://opensource.org/license/mit). You are free to use, modify, and distribute it, as long as you follow the terms of the license. You can find the full license text in the repository - [Full MIT license text](https://github.com/hueristiq/hq-lib-url-go/blob/main/LICENSE).
