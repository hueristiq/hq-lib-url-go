package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteSchemesToFile(t *testing.T) {
	t.Parallel()

	output := filepath.Join(t.TempDir(), "schemes.go")
	if err := writeSchemesToFile([]string{"http", "https"}, output); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Clean(output))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), `"http",`) || !strings.Contains(string(data), `"https",`) {
		t.Fatalf("generated file omits a scheme:\n%s", data)
	}
}
