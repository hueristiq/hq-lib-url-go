package parser_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func Test_URL_ComponentAccess(t *testing.T) {
	t.Parallel()

	p := hqgourlparser.New()

	parsed, err := p.Parse("https://user:pass@www.example.co.uk:8443/path/to/resource?a=1&b=2#section")
	require.NoError(t, err)

	assert.Equal(t, "https", parsed.Scheme)

	if assert.NotNil(t, parsed.User) {
		assert.Equal(t, "user", parsed.User.Username())

		password, ok := parsed.User.Password()

		assert.True(t, ok, "Expected the password to be set")
		assert.Equal(t, "pass", password)
	}

	assert.Equal(t, "www.example.co.uk:8443", parsed.Host)
	assert.Equal(t, "8443", parsed.Port())

	if assert.NotNil(t, parsed.Domain) {
		assert.Equal(t, "www.example.co.uk", parsed.Domain.String())
		assert.Equal(t, "www", parsed.Domain.Subdomain)
		assert.Equal(t, "example", parsed.Domain.SecondLevelDomain)
		assert.Equal(t, "co.uk", parsed.Domain.TopLevelDomain)
		assert.Equal(t, "example.co.uk", parsed.Domain.Apex())
	}

	assert.Equal(t, "/path/to/resource", parsed.Path)
	assert.Equal(t, "a=1&b=2", parsed.RawQuery)
	assert.Equal(t, "section", parsed.Fragment)
}

func Test_URL_ComponentAccess_IPHost(t *testing.T) {
	t.Parallel()

	p := hqgourlparser.New()

	parsed, err := p.Parse("http://192.168.0.1:8080/admin")
	require.NoError(t, err)

	require.Nil(t, parsed.Domain, "Expected IP literal hosts to carry no domain details")

	assert.Equal(t, "http", parsed.Scheme)
	assert.Equal(t, "192.168.0.1:8080", parsed.Host)
	assert.Equal(t, "8080", parsed.Port())
	assert.Equal(t, "/admin", parsed.Path)
}

func Test_URL_Normalize(t *testing.T) {
	t.Parallel()

	p := hqgourlparser.New()

	tests := []struct {
		name           string
		raw            string
		expectedString string
		expectedDomain *hqgourlparser.Domain
	}{
		{
			"lowercases host, drops default port, cleans path, sorts query",
			"https://WWW.EXAMPLE.COM:443/./a/../b/?b=2&a=1#Section",
			"https://www.example.com/b/?a=1&b=2#Section",
			&hqgourlparser.Domain{Subdomain: "www", SecondLevelDomain: "example", TopLevelDomain: "com"},
		},
		{
			"keeps non-default port",
			"http://example.com:8080/x",
			"http://example.com:8080/x",
			&hqgourlparser.Domain{Subdomain: "", SecondLevelDomain: "example", TopLevelDomain: "com"},
		},
		{
			"trims trailing root dot",
			"https://EXAMPLE.COM./x",
			"https://example.com/x",
			&hqgourlparser.Domain{Subdomain: "", SecondLevelDomain: "example", TopLevelDomain: "com"},
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
			&hqgourlparser.Domain{Subdomain: "", SecondLevelDomain: "example", TopLevelDomain: "com"},
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
	u := &hqgourlparser.URL{
		URL: &url.URL{
			Scheme: "http",
			Host:   "EXAMPLE.COM:bad",
			Path:   "/x",
		},
		Domain: &hqgourlparser.Domain{SecondLevelDomain: "example", TopLevelDomain: "com"},
	}

	u.Normalize()

	assert.Equal(t, "http://example.com:bad/x", u.String())
}

func Test_URL_Normalize_zeroValue(t *testing.T) {
	t.Parallel()

	var nilURL *hqgourlparser.URL

	nilURL.Normalize()
	(&hqgourlparser.URL{}).Normalize()
}
