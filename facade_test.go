package url_test

import (
	"strings"
	"testing"

	hqgourl "github.com/hueristiq/hq-lib-url-go"
)

func TestParseFacade(t *testing.T) {
	t.Parallel()

	parsed, err := hqgourl.Parse("https://example.com/path?q=1")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if parsed.Hostname() != "example.com" {
		t.Errorf("Hostname = %q", parsed.Hostname())
	}

	if parsed.Path != "/path" {
		t.Errorf("Path = %q", parsed.Path)
	}

	if _, err := hqgourl.Parse("http://[::1"); err == nil {
		t.Error("Parse: expected an error for an invalid URL")
	}
}

func TestExtractFacade(t *testing.T) {
	t.Parallel()

	matches, err := hqgourl.Extract("see https://example.com/a and http://example.org/b")
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	joined := strings.Join(matches, " ")

	if !strings.Contains(joined, "https://example.com/a") {
		t.Errorf("Extract = %q", joined)
	}

	if !strings.Contains(joined, "http://example.org/b") {
		t.Errorf("Extract = %q", joined)
	}

	empty, err := hqgourl.Extract("no urls here")
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	if len(empty) != 0 {
		t.Errorf("Extract = %q, want none", empty)
	}
}
