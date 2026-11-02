package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func TestNormalizeHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		host     string
		expected string
		wantErr  bool
	}{
		{name: "lowercase and root dot", host: "WWW.Example.COM.", expected: "www.example.com"},
		{name: "unicode to punycode", host: "münich.example", expected: "xn--mnich-kva.example"},
		{name: "IPv6 canonical", host: "2001:DB8::1", expected: "2001:db8::1"},
		{name: "underscore rejected", host: "_ldap._tcp.example.com", wantErr: true},
		{name: "empty", host: "", wantErr: true},
		{name: "space", host: "bad host", wantErr: true},
		{name: "leading hyphen", host: "-bad.example", wantErr: true},
		{name: "trailing hyphen", host: "bad-.example", wantErr: true},
		{name: "empty label", host: "bad..example", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			normalized, err := hqgourlparser.NormalizeHost(test.host)
			if test.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, normalized)
		})
	}
}

func TestValidateHost(t *testing.T) {
	t.Parallel()

	require.NoError(t, hqgourlparser.ValidateHost("example.com"))
	require.Error(t, hqgourlparser.ValidateHost(""))
}

func TestNormalizeDNSName(t *testing.T) {
	t.Parallel()

	normalized, err := hqgourlparser.NormalizeDNSName("_ldap._tcp.Example.COM.")
	require.NoError(t, err)
	assert.Equal(t, "_ldap._tcp.example.com", normalized)
}
