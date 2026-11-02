package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func TestURLNormalizeWith(t *testing.T) {
	t.Parallel()

	p := hqgourlparser.New(hqgourlparser.WithDefaultScheme("https"))

	parsed, err := p.Parse("HTTPS://Example.COM:443/b/../a/?b=2&a=1#section")
	require.NoError(t, err)

	parsed.Normalize()

	assert.Equal(t, "https", parsed.Scheme)
	assert.Equal(t, "example.com", parsed.Host)
	assert.Equal(t, "/a/", parsed.Path)
	assert.Equal(t, "a=1&b=2", parsed.RawQuery)
	assert.Equal(t, "section", parsed.Fragment)

	parsed, err = p.Parse("https://example.com/?b=2&a=1#section")
	require.NoError(t, err)

	parsed.NormalizeWith(hqgourlparser.NormalizeOptions{
		DropFragment:   true,
		KeepQueryOrder: true,
	})

	assert.Equal(t, "b=2&a=1", parsed.RawQuery)
	assert.Empty(t, parsed.Fragment)

	parsed, err = p.Parse("https://example.com/%2e%2e/x")
	require.NoError(t, err)

	parsed.NormalizeWith(hqgourlparser.NormalizeOptions{
		KeepDotSegments: true,
		KeepQueryOrder:  true,
	})

	assert.Equal(t, "https://example.com/%2e%2e/x", parsed.String())
}
