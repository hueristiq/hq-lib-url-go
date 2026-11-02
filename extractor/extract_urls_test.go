package extractor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
)

func TestExtractURLs(t *testing.T) {
	t.Parallel()

	urls, err := hqgourlextractor.ExtractURLs("see https://example.com/a and https://example.org/b")
	require.NoError(t, err)
	assert.Equal(t, []string{"https://example.com/a", "https://example.org/b"}, urls)

	urls, err = hqgourlextractor.ExtractURLs("https://example.com/a,https://example.org/b")
	require.NoError(t, err)
	assert.Equal(t, []string{"https://example.com/a", "https://example.org/b"}, urls)
}

func TestExtractURLsAll(t *testing.T) {
	t.Parallel()

	urls, err := hqgourlextractor.ExtractURLsAll(`https:\/\/example.com\/a?x=1`)
	require.NoError(t, err)
	assert.Contains(t, urls, "https://example.com/a?x=1")
}

func TestExtractCollapsedURLs(t *testing.T) {
	t.Parallel()

	urls, err := hqgourlextractor.ExtractCollapsedURLs(`const u = "https://" + "example.com/path";`)
	require.NoError(t, err)
	assert.Contains(t, urls, "https://example.com/path")
}

func TestExtractSubdomains(t *testing.T) {
	t.Parallel()

	subdomains, err := hqgourlextractor.ExtractSubdomains("see foo.example.com and bar.example.com.", "example.com")
	require.NoError(t, err)
	assert.Equal(t, []string{"foo.example.com", "bar.example.com"}, subdomains)

	_, err = hqgourlextractor.ExtractSubdomains("foo.example.com", "")
	require.Error(t, err)
}
