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
			"URL with unknown TLD decomposes positionally",
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
					TopLevelDomain:    "",
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
					TopLevelDomain:    "",
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
			"URL with punycode TLD is unrecognized",
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
					TopLevelDomain:    "",
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

func Test_Parser_Parse_doesNotMatchTLDSubstrings(t *testing.T) {
	t.Parallel()

	p := parser.New()

	tests := []struct {
		name           string
		raw            string
		expectedDomain *parser.Domain
	}{
		{
			// "or" is a substring of the "org" entry in the TLD list but not a TLD
			// itself; exact-match lookup must not confuse the two.
			"or is not a TLD",
			"https://example.or/",
			&parser.Domain{Subdomain: "", SecondLevelDomain: "example", TopLevelDomain: ""},
		},
		{
			"rg is not a TLD",
			"https://example.rg/",
			&parser.Domain{Subdomain: "", SecondLevelDomain: "example", TopLevelDomain: ""},
		},
		{
			"o is not a TLD",
			"https://example.o/",
			&parser.Domain{Subdomain: "", SecondLevelDomain: "example", TopLevelDomain: ""},
		},
		{
			// Positive control: "om" (Oman) is a real TLD and must match exactly.
			"om is a real TLD",
			"https://example.om/",
			&parser.Domain{Subdomain: "", SecondLevelDomain: "example", TopLevelDomain: "om"},
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

func Test_Parser_WithTLDs_Parse(t *testing.T) {
	t.Parallel()

	p := parser.New(parser.WithTLDs("custom"))

	tests := []struct {
		name              string
		URL               string
		expectedParsedURL *parser.URL
		expectedErr       bool
	}{
		{
			"URL with custom TLD",
			"https://example.custom/path",
			&parser.URL{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.custom",
					Path:   "/path",
				},
				Domain: &parser.Domain{
					Subdomain:         "",
					SecondLevelDomain: "example",
					TopLevelDomain:    "custom",
				},
			},
			false,
		},
		{
			"URL with default TLD is no longer recognized",
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
					TopLevelDomain:    "",
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

func Test_Parser_WithTLDs_Parse_multiLabelTLDWithoutParent(t *testing.T) {
	t.Parallel()

	// "uk" is not configured, but "co.uk" is: the scan must continue past the
	// non-matching "uk" suffix and recognize the longer configured suffix.
	p := parser.New(parser.WithTLDs("co.uk"))

	actualParsedURL, err := p.Parse("https://example.co.uk/path")

	require.NoError(t, err)
	assert.Equal(t, &parser.URL{
		URL: &url.URL{
			Scheme: "https",
			Host:   "example.co.uk",
			Path:   "/path",
		},
		Domain: &parser.Domain{
			Subdomain:         "",
			SecondLevelDomain: "example",
			TopLevelDomain:    "co.uk",
		},
	}, actualParsedURL)
}

func Test_URL_Key(t *testing.T) {
	t.Parallel()

	p := parser.New()

	parsed, err := p.Parse("https://user:pass@www.example.co.uk:8443/path/to/resource?a=1&b=2#section")
	require.NoError(t, err)

	tests := []struct {
		key           string
		expectedValue string
	}{
		{"scheme", "https"},
		{"user", "user"},
		{"pass", "pass"},
		{"host", "www.example.co.uk:8443"},
		{"port", "8443"},
		{"domain", "www.example.co.uk"},
		{"subdomain", "www"},
		{"sld", "example"},
		{"root", "example"},
		{"tld", "co.uk"},
		{"apex", "example.co.uk"},
		{"path", "/path/to/resource"},
		{"query", "a=1&b=2"},
		{"fragment", "section"},
		{"bogus", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expectedValue, parsed.Key(tt.key))
		})
	}
}

func Test_URL_Key_IPHost(t *testing.T) {
	t.Parallel()

	p := parser.New()

	parsed, err := p.Parse("http://192.168.0.1:8080/admin")
	require.NoError(t, err)

	require.Nil(t, parsed.Domain)

	tests := []struct {
		key           string
		expectedValue string
	}{
		{"scheme", "http"},
		{"host", "192.168.0.1:8080"},
		{"port", "8080"},
		{"domain", ""},
		{"subdomain", ""},
		{"sld", ""},
		{"tld", ""},
		{"apex", ""},
		{"path", "/admin"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expectedValue, parsed.Key(tt.key))
		})
	}
}

func Test_URL_Normalize(t *testing.T) {
	t.Parallel()

	p := parser.New()

	tests := []struct {
		name           string
		raw            string
		expectedString string
		expectedDomain *parser.Domain
	}{
		{
			"lowercases host, drops default port, cleans path, sorts query",
			"https://WWW.EXAMPLE.COM:443/./a/../b/?b=2&a=1#Section",
			"https://www.example.com/b/?a=1&b=2#Section",
			&parser.Domain{Subdomain: "www", SecondLevelDomain: "example", TopLevelDomain: "com"},
		},
		{
			"keeps non-default port",
			"http://example.com:8080/x",
			"http://example.com:8080/x",
			&parser.Domain{Subdomain: "", SecondLevelDomain: "example", TopLevelDomain: "com"},
		},
		{
			"trims trailing root dot",
			"https://EXAMPLE.COM./x",
			"https://example.com/x",
			&parser.Domain{Subdomain: "", SecondLevelDomain: "example", TopLevelDomain: "com"},
		},
		{
			"drops default port for IPv6 and restores brackets",
			"http://[2001:DB8::1]:80/a",
			"http://[2001:db8::1]/a",
			nil,
		},
		{
			"keeps unparseable query untouched",
			"https://example.com/?a=%zz",
			"https://example.com/?a=%zz",
			&parser.Domain{Subdomain: "", SecondLevelDomain: "example", TopLevelDomain: "com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := p.Parse(tt.raw)
			require.NoError(t, err)

			parsed.Normalize()

			assert.Equal(t, tt.expectedString, parsed.String())
			assert.Equal(t, tt.expectedDomain, parsed.Domain)
		})
	}
}

func Test_URL_Normalize_malformedPort(t *testing.T) {
	t.Parallel()

	// net/url rejects a malformed port when parsing, so this case builds the URL by
	// hand: Normalize must preserve the port verbatim rather than silently dropping it.
	u := &parser.URL{
		URL: &url.URL{
			Scheme: "http",
			Host:   "EXAMPLE.COM:bad",
			Path:   "/x",
		},
		Domain: &parser.Domain{SecondLevelDomain: "example", TopLevelDomain: "com"},
	}

	u.Normalize()

	assert.Equal(t, "http://example.com:bad/x", u.String())
}

func Test_InferScheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		host           string
		expectedScheme string
	}{
		{"example.com", "https"},
		{"example.com:8443", "https"},
		{"8.8.8.8", "https"},
		{"EXAMPLE.COM.", "https"},
		{"localhost", "http"},
		{"localhost:8080", "http"},
		{"foo.localhost", "http"},
		{"host.local", "http"},
		{"host.test", "http"},
		{"host.invalid", "http"},
		{"127.0.0.1", "http"},
		{"127.0.0.1:8080", "http"},
		{"::1", "http"},
		{"10.0.0.1", "http"},
		{"192.168.1.1", "http"},
		{"169.254.0.1", "http"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expectedScheme, parser.InferScheme(tt.host))
		})
	}
}

func Test_Domain_String_And_Apex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		domain       parser.Domain
		expectedFull string
		expectedApex string
	}{
		{
			"full domain",
			parser.Domain{Subdomain: "www", SecondLevelDomain: "example", TopLevelDomain: "com"},
			"www.example.com",
			"example.com",
		},
		{
			"without subdomain",
			parser.Domain{SecondLevelDomain: "example", TopLevelDomain: "com"},
			"example.com",
			"example.com",
		},
		{
			"single label",
			parser.Domain{SecondLevelDomain: "localhost"},
			"localhost",
			"localhost",
		},
		{
			"suffix only",
			parser.Domain{TopLevelDomain: "co.uk"},
			"co.uk",
			"co.uk",
		},
		{
			"zero domain",
			parser.Domain{},
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expectedFull, tt.domain.String())
			assert.Equal(t, tt.expectedApex, tt.domain.Apex())
		})
	}
}
