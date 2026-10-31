// Command pslgen regenerates internal/psl/rules_gen.go from the Public Suffix List.
//
// It fetches the list from publicsuffix.org, converts internationalized rules to
// A-labels, and emits the embedded rule table used by the internal/psl package.
// Run it from
// the module root through go generate.
package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/format"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/idna"
)

const (
	// pslURL is the canonical location of the Public Suffix List source.
	pslURL = "https://publicsuffix.org/list/public_suffix_list.dat"

	// httpTimeout bounds a single download of the Public Suffix List.
	httpTimeout = 60 * time.Second
)

// rule is one Public Suffix List rule as emitted to the generated table.
type rule struct {
	kind    string
	value   string
	labels  int
	private bool
}

// main parses the command-line flags and regenerates the rule table.
func main() {
	output := flag.String("output", "", "path of the generated Go file")

	flag.Parse()

	if *output == "" {
		log.Fatal("output file path is required: use -output to specify it")
	}

	if err := run(*output); err != nil {
		log.Fatal(err)
	}
}

// run fetches the Public Suffix List, renders the rule table, and writes it to output.
//
// Parameters:
//   - output (string): the path of the generated Go file.
//
// Returns:
//   - err (error): non-nil when the list cannot be fetched or the file cannot be
//     written.
func run(output string) (err error) {
	rules, version, commit, err := fetchRules()
	if err != nil {
		return err
	}

	data, err := render(rules, version, commit)
	if err != nil {
		return err
	}

	if err = os.WriteFile(output, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", output, err)
	}

	log.Printf("generated %s with %d rules (PSL %s, commit %s)", output, len(rules), version, commit)

	return nil
}

// fetchRules downloads the Public Suffix List and parses its rules.
//
// Internationalized rules are emitted twice: once in their original form and once as
// an A-label, so hosts in either notation match.
//
// Returns:
//   - rules ([]rule): the parsed rules, sorted by value and kind.
//   - version (string): the list's VERSION header, when present.
//   - commit (string): the list's COMMIT header, when present.
//   - err (error): non-nil when the list cannot be fetched or read.
func fetchRules() (rules []rule, version, commit string, err error) {
	client := &http.Client{Timeout: httpTimeout}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, pslURL, http.NoBody)
	if err != nil {
		return nil, "", "", fmt.Errorf("build PSL request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, "", "", fmt.Errorf("fetch PSL: %w", err)
	}

	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, "", "", fmt.Errorf("fetch PSL: unexpected status %s", res.Status)
	}

	seen := make(map[string]struct{})

	private := false

	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		case strings.HasPrefix(line, "// VERSION:"):
			version = strings.TrimSpace(strings.TrimPrefix(line, "// VERSION:"))

			continue
		case strings.HasPrefix(line, "// COMMIT:"):
			commit = strings.TrimSpace(strings.TrimPrefix(line, "// COMMIT:"))

			continue
		case strings.HasPrefix(line, "// ===BEGIN PRIVATE DOMAINS==="):
			private = true

			continue
		case strings.HasPrefix(line, "// ===END PRIVATE DOMAINS==="):
			private = false

			continue
		case line == "", strings.HasPrefix(line, "//"):
			continue
		}

		kind := "normal"
		value := line

		switch {
		case strings.HasPrefix(value, "*."):
			kind = "wildcard"
			value = strings.TrimPrefix(value, "*.")
		case strings.HasPrefix(value, "!"):
			kind = "exception"
			value = strings.TrimPrefix(value, "!")
		}

		value = strings.ToLower(strings.TrimSuffix(value, "."))

		if value == "" || value == "*" {
			continue
		}

		values := []string{value}

		if !isASCII(value) {
			asciiValue, convertErr := idna.ToASCII(value)
			if convertErr != nil {
				return nil, "", "", fmt.Errorf("convert %q: %w", line, convertErr)
			}

			if asciiValue != value {
				values = append(values, asciiValue)
			}
		}

		for _, candidate := range values {
			key := fmt.Sprintf("%t:%s:%s", private, kind, candidate)
			if _, ok := seen[key]; ok {
				continue
			}

			seen[key] = struct{}{}

			labels := len(strings.Split(candidate, "."))
			if kind == "wildcard" {
				labels++
			}

			rules = append(rules, rule{kind: kind, value: candidate, labels: labels, private: private})
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, "", "", fmt.Errorf("read PSL: %w", err)
	}

	sort.Slice(rules, func(i, j int) bool {
		if rules[i].value == rules[j].value {
			return rules[i].kind < rules[j].kind
		}

		return rules[i].value < rules[j].value
	})

	return rules, version, commit, nil
}

// render formats the rules as the source of the generated rules file.
//
// Parameters:
//   - rules ([]rule): the rules to render.
//   - version (string): the Public Suffix List VERSION header.
//   - commit (string): the Public Suffix List COMMIT header.
//
// Returns:
//   - data ([]byte): the gofmt-ed source of the generated file.
//   - err (error): non-nil when a rule kind is unknown or formatting fails.
func render(rules []rule, version, commit string) (data []byte, err error) {
	kindNames := map[string]string{
		"normal":    "pslNormal",
		"wildcard":  "pslWildcard",
		"exception": "pslException",
	}

	var buf bytes.Buffer

	buf.WriteString("// Code generated by cmd/pslgen; DO NOT EDIT.\n")
	fmt.Fprintf(&buf, "// PSL VERSION: %s\n", version)
	fmt.Fprintf(&buf, "// PSL COMMIT: %s\n", commit)
	buf.WriteString("\npackage psl\n\n")
	buf.WriteString("var pslRules = [...]Rule{\n")

	for _, r := range rules {
		name, ok := kindNames[r.kind]
		if !ok {
			return nil, fmt.Errorf("unknown rule kind %q", r.kind)
		}

		fmt.Fprintf(&buf, "\t{%s, %q, %d, %t},\n", name, r.value, r.labels, r.private)
	}

	buf.WriteString("}\n")

	data, err = format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format rules: %w", err)
	}

	return data, nil
}

// isASCII reports whether value contains only ASCII bytes.
//
// Parameters:
//   - value (string): the string to inspect.
//
// Returns:
//   - ascii (bool): true when every byte of value is below 0x80.
func isASCII(value string) (ascii bool) {
	for i := range len(value) {
		if value[i] >= 0x80 {
			return false
		}
	}

	return true
}
