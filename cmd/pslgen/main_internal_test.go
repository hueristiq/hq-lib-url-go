package main

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	t.Parallel()

	rules := []rule{
		{kind: "normal", value: "com", labels: 1},
		{kind: "wildcard", value: "ck", labels: 2, private: true},
	}

	data, err := render(rules, "version", "commit")
	if err != nil {
		t.Fatal(err)
	}

	text := string(data)
	if !strings.Contains(text, `{pslNormal, "com", 1, false}`) || !strings.Contains(text, `{pslWildcard, "ck", 2, true}`) {
		t.Fatalf("render() omitted a rule:\n%s", text)
	}

	if _, err := render([]rule{{kind: "unknown"}}, "", ""); err == nil {
		t.Fatal("expected an unknown rule-kind error")
	}
}
