package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteTLDsToFile(t *testing.T) {
	t.Parallel()

	output := filepath.Join(t.TempDir(), "tlds.go")
	if err := writeTLDsToFile([]string{"com", "co.uk"}, output); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Clean(output))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), `"com",`) || !strings.Contains(string(data), `"co.uk",`) {
		t.Fatalf("generated file omits a TLD:\n%s", data)
	}
}
