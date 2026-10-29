package parser_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func Test_Parser_Parse(t *testing.T) {
	t.Parallel()

	p := hqgourlparser.New()

	tests := []struct {
		name           string
		raw            string
		expectedParsed *hqgourlparser.URL
		expectedErr    bool
	}{
		{
			"URL",
			"https://example.com/path",
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.com",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.co.uk",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.bo.nordland.no",
					Path:   "/",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.invalidtld",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.invalidtld",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.local",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.中国",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.xn--p1ai",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.com:8080",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "http",
					Host:   "localhost:8080",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "co.uk",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.ck",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "test.ck",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "b.test.ck",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "blogspot.com",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "foo.blogspot.com",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "s3.amazonaws.com",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "bucket.s3.amazonaws.com",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "WWW.EXAMPLE.COM",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com.",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
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
			&hqgourlparser.URL{
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
			&hqgourlparser.URL{},
			true,
		},
		{
			"Invalid URL 1",
			"https://example.com/%invalid",
			&hqgourlparser.URL{},
			true,
		},
		{
			"URL with percent-encoded CR in path",
			"https://example.com/w%0d%2e/",
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme:  "https",
					Host:    "example.com",
					Path:    "/w\r./",
					RawPath: "/w%0d%2e/",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme:   "",
					Host:     "",
					Path:     "example.com/with/path",
					RawQuery: "some'param=`'+OR+ORDER+BY+1--",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/a/b/c/../c",
				},
				Domain: &hqgourlparser.Domain{
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

	p := hqgourlparser.New(hqgourlparser.WithDefaultScheme("https"))

	tests := []struct {
		name              string
		URL               string
		expectedParsedURL *hqgourlparser.URL
		expectedErr       bool
	}{
		{
			"URL without scheme",
			"example.com/path",
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
					Path:   "/path",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "mailto",
					Opaque: "user@example.com",
				},
				Domain: &hqgourlparser.Domain{},
			},
			false,
		},
		{
			// A ":" in the first path segment followed by non-digits stays opaque.
			"Opaque URI with non-numeric remainder",
			"foo:bar",
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "foo",
					Opaque: "bar",
				},
				Domain: &hqgourlparser.Domain{},
			},
			false,
		},
		{
			// A scheme-less "host:port" pair (digits after the colon) is a host: the
			// default scheme is prepended and the port is kept.
			"URL without scheme with host and port",
			"example.com:8080/api",
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com:8080",
					Path:   "/api",
				},
				Domain: &hqgourlparser.Domain{
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
			&hqgourlparser.URL{
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
			&hqgourlparser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "localhost:8080",
				},
				Domain: &hqgourlparser.Domain{
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

	p := hqgourlparser.New(hqgourlparser.WithIgnorePrivate())

	tests := []struct {
		name           string
		raw            string
		expectedDomain *hqgourlparser.Domain
	}{
		{
			"private suffix as the whole host",
			"https://blogspot.com/path",
			&hqgourlparser.Domain{SecondLevelDomain: "blogspot", TopLevelDomain: "com"},
		},
		{
			"host under a private suffix",
			"https://foo.blogspot.com/path",
			&hqgourlparser.Domain{Subdomain: "foo", SecondLevelDomain: "blogspot", TopLevelDomain: "com"},
		},
		{
			"multi-label private suffix as the whole host",
			"https://s3.amazonaws.com/path",
			&hqgourlparser.Domain{Subdomain: "s3", SecondLevelDomain: "amazonaws", TopLevelDomain: "com"},
		},
		{
			"host under a multi-label private suffix",
			"https://bucket.s3.amazonaws.com/path",
			&hqgourlparser.Domain{Subdomain: "bucket.s3", SecondLevelDomain: "amazonaws", TopLevelDomain: "com"},
		},
		{
			"unknown TLD still uses the default rule",
			"https://example.invalidtld/path",
			&hqgourlparser.Domain{SecondLevelDomain: "example", TopLevelDomain: "invalidtld"},
		},
		{
			"public multi-label suffix is unchanged",
			"https://www.example.co.uk/path",
			&hqgourlparser.Domain{Subdomain: "www", SecondLevelDomain: "example", TopLevelDomain: "co.uk"},
		},
		{
			"public exception rule is unchanged",
			"https://www.ck/path",
			&hqgourlparser.Domain{SecondLevelDomain: "www", TopLevelDomain: "ck"},
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

	var p hqgourlparser.Parser

	parsed, err := p.Parse("https://a.b.example/path")
	require.NoError(t, err)

	assert.Equal(t, &hqgourlparser.Domain{Subdomain: "a", SecondLevelDomain: "b"}, parsed.Domain)
}

func TestParse_Lenient(t *testing.T) {
	t.Parallel()

	lenient := hqgourlparser.New(
		hqgourlparser.WithDefaultScheme("http"),
		hqgourlparser.WithLenient(),
	)

	t.Run("invalid escape", func(t *testing.T) {
		t.Parallel()

		parsed, err := lenient.Parse("https://scanme.sh/%invalid")

		require.NoError(t, err)
		require.NotNil(t, parsed)

		assert.Equal(t, "https", parsed.Scheme)
		assert.Equal(t, "scanme.sh", parsed.Host)
		assert.Equal(t, "/%invalid", parsed.Path)
		assert.Empty(t, parsed.RawPath, "Expected net/url to encode the repaired path directly")
		assert.Equal(t, "/%25invalid", parsed.EscapedPath())
	})

	t.Run("preserved encoded path", func(t *testing.T) {
		t.Parallel()

		parsed, err := lenient.Parse("https://scanme.sh/w%0d%2e/")

		require.NoError(t, err)
		require.NotNil(t, parsed)

		assert.Equal(t, "https", parsed.Scheme)
		assert.Equal(t, "scanme.sh", parsed.Host)
		assert.Equal(t, "/w\r./", parsed.Path)
		assert.Equal(t, "/w%0d%2e/", parsed.RawPath)
		assert.Equal(t, "/w%0d%2e/", parsed.EscapedPath())
	})

	t.Run("scheme-less host with raw query", func(t *testing.T) {
		t.Parallel()

		const rawQuery = "some'param=`'+OR+ORDER+BY+1--"

		parsed, err := lenient.Parse("scanme.sh/with/path?" + rawQuery)

		require.NoError(t, err)
		require.NotNil(t, parsed)

		assert.Equal(t, "http", parsed.Scheme)
		assert.Equal(t, "scanme.sh", parsed.Host)
		assert.Equal(t, "/with/path", parsed.Path)
		assert.Equal(t, rawQuery, parsed.RawQuery)
	})

	t.Run("valid escape is not double encoded", func(t *testing.T) {
		t.Parallel()

		parsed, err := lenient.Parse("https://scanme.sh/a%2Fb")

		require.NoError(t, err)
		require.NotNil(t, parsed)

		assert.Equal(t, "/a/b", parsed.Path)
		assert.Equal(t, "/a%2Fb", parsed.RawPath)
		assert.Equal(t, "/a%2Fb", parsed.EscapedPath())
	})

	t.Run("invalid host still fails", func(t *testing.T) {
		t.Parallel()

		parsed, err := lenient.Parse("http://exa mple.com")

		require.Error(t, err)
		assert.Nil(t, parsed)
	})

	t.Run("strict default still rejects", func(t *testing.T) {
		t.Parallel()

		strict := hqgourlparser.New(hqgourlparser.WithDefaultScheme("http"))

		parsed, err := strict.Parse("https://scanme.sh/%invalid")

		require.Error(t, err)
		assert.Nil(t, parsed)
	})
}

func Test_Parser_New_IgnoresNilOptions(t *testing.T) {
	t.Parallel()

	p := hqgourlparser.New(nil)
	require.NotNil(t, p)
}

// ExampleParse demonstrates parsing a URL and reading its decomposed domain.
func ExampleParse() {
	p := hqgourlparser.New(hqgourlparser.WithDefaultScheme("https"))

	parsed, err := p.Parse("scanme.sh/path")
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(parsed.Scheme)
	fmt.Println(parsed.Domain.Subdomain)
	fmt.Println(parsed.Domain.SecondLevelDomain)
	fmt.Println(parsed.Domain.TopLevelDomain)
	// Output:
	// https
	//
	// scanme
	// sh
}

// FuzzParse exercises Parse over arbitrary inputs: no input may panic, and a
// rejected URL must return an error rather than a partial result.
func FuzzParse(f *testing.F) {
	f.Add("https://example.com/path")
	f.Add("example.com")
	f.Add("")
	f.Add("http://[::1]:8080/x")
	f.Add("https://user:pass@www.example.co.uk:8443/a?b=c#d")
	f.Add("%zz")

	f.Fuzz(func(t *testing.T, raw string) {
		p := hqgourlparser.New()

		parsed, err := p.Parse(raw)
		if err != nil {
			return
		}

		if parsed == nil || parsed.URL == nil {
			t.Fatal("Expected a non-nil parsed URL on success")
		}

		// The domain must be consistent with the host: IP literals carry nil,
		// empty hosts carry a zero Domain.
		emptyHost := parsed.Hostname() == "" && parsed.Domain != nil &&
			(parsed.Domain.Subdomain != "" || parsed.Domain.SecondLevelDomain != "" || parsed.Domain.TopLevelDomain != "")

		if emptyHost {
			t.Fatal("Expected an empty-host URL to carry no domain components")
		}
	})
}
