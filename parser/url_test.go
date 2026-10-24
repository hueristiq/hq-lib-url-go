package parser_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hueristiq/hq-lib-url-go/parser"
)

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

func Test_URL_Key_zeroValue(t *testing.T) {
	t.Parallel()

	var nilURL *parser.URL

	assert.Empty(t, nilURL.Key("scheme"))
	assert.Empty(t, (&parser.URL{}).Key("scheme"))
}

func Test_URL_Normalize_zeroValue(t *testing.T) {
	t.Parallel()

	var nilURL *parser.URL

	nilURL.Normalize()
	(&parser.URL{}).Normalize()
}
