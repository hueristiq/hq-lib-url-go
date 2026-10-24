package parser_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hueristiq/hq-lib-url-go/parser"
)

func Test_Parser_Parse(t *testing.T) {
	t.Parallel()

	p := parser.New()

	tests := []struct {
		name           string
		raw            string
		expectedParsed *parser.URL
		expectedErr    bool
	}{
		{
			"URL",
			"https://example.com/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
		{
			"URL with subdomain",
			"https://www.example.com/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.com",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "www",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
		{
			"URL with multi-label TLD",
			"https://www.example.co.uk/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.co.uk",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "www",
					SecondLevelDomain: "example",
					TopLevelDomain:    "co.uk",
				},
			},
			false,
		},
		{
			// "bo.nordland.no" is a recognized suffix while "nordland.no" is not: the
			// scan must continue past the non-matching suffix and keep the longest match.
			"URL with nested multi-label TLD",
			"https://www.bo.nordland.no/",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.bo.nordland.no",
					Path:   "/",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "www",
					TopLevelDomain:    "bo.nordland.no",
				},
			},
			false,
		},
		{
			"URL with unknown TLD uses the default rule",
			"https://www.example.invalidtld/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.invalidtld",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "www",
					SecondLevelDomain: "example",
					TopLevelDomain:    "invalidtld",
				},
			},
			false,
		},
		{
			"URL with unknown TLD and no subdomain",
			"https://example.invalidtld/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.invalidtld",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "invalidtld",
				},
			},
			false,
		},
		{
			"URL with pseudo TLD",
			"https://www.example.local/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.local",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "www",
					SecondLevelDomain: "example",
					TopLevelDomain:    "local",
				},
			},
			false,
		},
		{
			"URL with unicode TLD",
			"https://example.中国/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.中国",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "中国",
				},
			},
			false,
		},
		{
			"URL with punycode TLD is recognized",
			"https://example.xn--p1ai/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.xn--p1ai",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "xn--p1ai",
				},
			},
			false,
		},
		{
			"URL with port",
			"https://www.example.com:8080/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.com:8080",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "www",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
		{
			"URL with single-label host",
			"http://localhost:8080/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "http",
					Host:   "localhost:8080",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "localhost",
					TopLevelDomain:    "",
				},
			},
			false,
		},
		{
			"URL with whole-host recognized suffix",
			"https://co.uk/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "co.uk",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "",
					TopLevelDomain:    "co.uk",
				},
			},
			false,
		},
		// Public suffix rules from the ICANN section: the exception rule makes
		// "www.ck" registrable under "ck", while the wildcard rule makes "test.ck"
		// itself a public suffix.
		{
			"URL with public suffix exception rule",
			"https://www.ck/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.ck",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "www",
					TopLevelDomain:    "ck",
				},
			},
			false,
		},
		{
			"URL with public wildcard suffix",
			"https://test.ck/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "test.ck",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "",
					TopLevelDomain:    "test.ck",
				},
			},
			false,
		},
		{
			"URL with public wildcard suffix and subdomain",
			"https://b.test.ck/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "b.test.ck",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "b",
					TopLevelDomain:    "test.ck",
				},
			},
			false,
		},
		// Private-section suffixes are public suffixes too: the whole host lands
		// in TopLevelDomain, and a label before it becomes SecondLevelDomain.
		{
			"URL with private suffix",
			"https://blogspot.com/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "blogspot.com",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "",
					TopLevelDomain:    "blogspot.com",
				},
			},
			false,
		},
		{
			"URL with private suffix and subdomain",
			"https://foo.blogspot.com/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "foo.blogspot.com",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "foo",
					TopLevelDomain:    "blogspot.com",
				},
			},
			false,
		},
		{
			"URL with multi-label private suffix",
			"https://s3.amazonaws.com/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "s3.amazonaws.com",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "",
					TopLevelDomain:    "s3.amazonaws.com",
				},
			},
			false,
		},
		{
			"URL with multi-label private suffix and subdomain",
			"https://bucket.s3.amazonaws.com/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "bucket.s3.amazonaws.com",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "bucket",
					TopLevelDomain:    "s3.amazonaws.com",
				},
			},
			false,
		},
		{
			"URL with uppercase host",
			"https://WWW.EXAMPLE.COM/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "WWW.EXAMPLE.COM",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "www",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
		{
			"URL with trailing root dot",
			"https://example.com./path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com.",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
		{
			"URL with IPv4",
			"http://192.168.0.1/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "http",
					Host:   "192.168.0.1",
					Path:   "/path",
				},
				Domain: nil,
			},
			false,
		},
		{
			"URL with IPv6",
			"https://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:17000/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:17000",
					Path:   "/path",
				},
				Domain: nil,
			},
			false,
		},
		{
			"Invalid URL 0",
			"://example.com",
			&parser.URL{},
			true,
		},
		{
			"Invalid URL 1",
			"https://example.com/%invalid",
			&parser.URL{},
			true,
		},
		{
			"URL with percent-encoded CR in path",
			"https://example.com/w%0d%2e/",
			&parser.URL{
				URL: &url.URL{
					Scheme:  "https",
					Host:    "example.com",
					Path:    "/w\r./",
					RawPath: "/w%0d%2e/",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
		{
			"Scheme-less URL with query",
			"example.com/with/path?some'param=`'+OR+ORDER+BY+1--",
			&parser.URL{
				URL: &url.URL{
					Scheme:   "",
					Host:     "",
					Path:     "example.com/with/path",
					RawQuery: "some'param=`'+OR+ORDER+BY+1--",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "",
					TopLevelDomain:    "",
				},
			},
			false,
		},
		{
			"URL with relative dot segments in path",
			"https://example.com/a/b/c/../c",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/a/b/c/../c",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actualParsedURL, err := p.Parse(tt.raw)

			if tt.expectedErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedParsed, actualParsedURL)
		})
	}
}

func Test_Parser_WithDefaultScheme_Parse(t *testing.T) {
	t.Parallel()

	p := parser.New(parser.WithDefaultScheme("https"))

	tests := []struct {
		name              string
		URL               string
		expectedParsedURL *parser.URL
		expectedErr       bool
	}{
		{
			"URL without scheme",
			"example.com/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
		{
			"URL with ://",
			"://example.com/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
		{
			"network-path reference",
			"//example.com/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
		{
			"URL with scheme",
			"http://example.com/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
		{
			// A ":" in the first path segment marks an opaque URI with an explicit
			// scheme: the default scheme must not be prepended.
			"Opaque URI with explicit scheme",
			"mailto:user@example.com",
			&parser.URL{
				URL: &url.URL{
					Scheme: "mailto",
					Opaque: "user@example.com",
				},
				Domain: &parser.Domain{},
			},
			false,
		},
		{
			// A ":" in the first path segment followed by non-digits stays opaque.
			"Opaque URI with non-numeric remainder",
			"foo:bar",
			&parser.URL{
				URL: &url.URL{
					Scheme: "foo",
					Opaque: "bar",
				},
				Domain: &parser.Domain{},
			},
			false,
		},
		{
			// A scheme-less "host:port" pair (digits after the colon) is a host: the
			// default scheme is prepended and the port is kept.
			"URL without scheme with host and port",
			"example.com:8080/api",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com:8080",
					Path:   "/api",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "com",
				},
			},
			false,
		},
		{
			"URL without scheme with IP and port",
			"127.0.0.1:8080/x",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "127.0.0.1:8080",
					Path:   "/x",
				},
			},
			false,
		},
		{
			"URL without scheme with single-label host and port",
			"localhost:8080",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "localhost:8080",
				},
				Domain: &parser.Domain{
					SecondLevelDomain: "localhost",
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actualParsedURL, err := p.Parse(tt.URL)

			if tt.expectedErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedParsedURL, actualParsedURL)
		})
	}
}

func Test_Parser_WithIgnorePrivate(t *testing.T) {
	t.Parallel()

	p := parser.New(parser.WithIgnorePrivate())

	tests := []struct {
		name           string
		raw            string
		expectedDomain *parser.Domain
	}{
		{
			"private suffix as the whole host",
			"https://blogspot.com/path",
			&parser.Domain{SecondLevelDomain: "blogspot", TopLevelDomain: "com"},
		},
		{
			"host under a private suffix",
			"https://foo.blogspot.com/path",
			&parser.Domain{Subdomain: "foo", SecondLevelDomain: "blogspot", TopLevelDomain: "com"},
		},
		{
			"multi-label private suffix as the whole host",
			"https://s3.amazonaws.com/path",
			&parser.Domain{Subdomain: "s3", SecondLevelDomain: "amazonaws", TopLevelDomain: "com"},
		},
		{
			"host under a multi-label private suffix",
			"https://bucket.s3.amazonaws.com/path",
			&parser.Domain{Subdomain: "bucket.s3", SecondLevelDomain: "amazonaws", TopLevelDomain: "com"},
		},
		{
			"unknown TLD still uses the default rule",
			"https://example.invalidtld/path",
			&parser.Domain{SecondLevelDomain: "example", TopLevelDomain: "invalidtld"},
		},
		{
			"public multi-label suffix is unchanged",
			"https://www.example.co.uk/path",
			&parser.Domain{Subdomain: "www", SecondLevelDomain: "example", TopLevelDomain: "co.uk"},
		},
		{
			"public exception rule is unchanged",
			"https://www.ck/path",
			&parser.Domain{SecondLevelDomain: "www", TopLevelDomain: "ck"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := p.Parse(tt.raw)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedDomain, parsed.Domain)
		})
	}
}

func Test_Parser_zeroValue(t *testing.T) {
	t.Parallel()

	var p parser.Parser

	parsed, err := p.Parse("https://a.b.example/path")
	require.NoError(t, err)

	assert.Equal(t, &parser.Domain{Subdomain: "a", SecondLevelDomain: "b"}, parsed.Domain)
}
