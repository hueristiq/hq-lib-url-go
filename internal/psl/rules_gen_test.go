package psl_test

import (
	"testing"

	"github.com/hueristiq/hq-lib-url-go/internal/psl"
)

func TestGeneratedRules(t *testing.T) {
	t.Parallel()

	rule := psl.DefaultList().Find("www.example.co.uk")
	if rule == nil || rule.Suffix() != "co.uk" {
		t.Fatalf("Find() = %#v; want the co.uk rule", rule)
	}
}
