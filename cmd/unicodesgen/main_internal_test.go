package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode"
)

func TestVisit(t *testing.T) {
	t.Parallel()

	table := &unicode.RangeTable{
		R16: []unicode.Range16{{Lo: 1, Hi: 3, Stride: 1}},
		R32: []unicode.Range32{{Lo: 10, Hi: 12, Stride: 2}},
	}

	var visited []rune

	visit(table, func(value rune) { visited = append(visited, value) })

	if !slices.Equal(visited, []rune{1, 2, 3, 10, 12}) {
		t.Fatalf("visit() = %v", visited)
	}
}

func TestWriteUnicode(t *testing.T) {
	t.Parallel()

	output := filepath.Join(t.TempDir(), "unicodes.go")
	if err := writeUnicode(output); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Clean(output))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "const AllowedUcsChar") || !strings.Contains(string(data), "const AllowedUcsCharMinusPunc") {
		t.Fatalf("generated file omits a character set:\n%s", data)
	}
}
