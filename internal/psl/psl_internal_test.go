package psl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRule_Decompose(t *testing.T) {
	t.Parallel()

	left, suffix := (&Rule{kind: pslNormal, value: "com", labels: 1}).Decompose("example.com")
	assert.Equal(t, "example", left)
	assert.Equal(t, "com", suffix)

	left, suffix = (&Rule{kind: pslWildcard, value: "ck", labels: 2}).Decompose("a.b.ck")
	assert.Equal(t, "a", left)
	assert.Equal(t, "b.ck", suffix)

	left, suffix = (&Rule{kind: pslException, value: "www.ck", labels: 2}).Decompose("www.ck")
	assert.Equal(t, "www", left)
	assert.Equal(t, "ck", suffix)
}

func TestRule_DecomposeUnmatched(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rule *Rule
		host string
	}{
		{"default rule without a dot", defaultRule, "localhost"},
		{"normal host equal to the suffix", &Rule{kind: pslNormal, value: "com", labels: 1}, "com"},
		{"wildcard host equal to the suffix", &Rule{kind: pslWildcard, value: "ck", labels: 2}, "ck"},
		{"wildcard without a label before the suffix", &Rule{kind: pslWildcard, value: "ck", labels: 2}, "b.ck"},
		{"exception value without a dot", &Rule{kind: pslException, value: "foo", labels: 1}, "x.foo"},
		{"exception host equal to the suffix", &Rule{kind: pslException, value: "www.ck", labels: 2}, "ck"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			left, suffix := tt.rule.Decompose(tt.host)
			assert.Empty(t, left)
			assert.Empty(t, suffix)
		})
	}
}

func TestList_Find(t *testing.T) {
	t.Parallel()

	list := newList([]Rule{
		{kind: pslNormal, value: "com", labels: 1},
		{kind: pslNormal, value: "co.uk", labels: 2},
		{kind: pslWildcard, value: "ck", labels: 2},
		{kind: pslException, value: "www.ck", labels: 2},
	}, defaultRule, true)

	tests := []struct {
		name   string
		host   string
		left   string
		suffix string
	}{
		{"normal", "example.com", "example", "com"},
		{"multi-label normal", "example.co.uk", "example", "co.uk"},
		{"wildcard", "a.b.ck", "a", "b.ck"},
		{"exception", "www.ck", "www", "ck"},
		{"default rule", "example.invalid", "example", "invalid"},
		{"wildcard label alone falls through", "ck", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := list.Find(tt.host)
			left, suffix := rule.Decompose(tt.host)
			assert.Equal(t, tt.left, left)
			assert.Equal(t, tt.suffix, suffix)
		})
	}
}

func TestList_FindNoDefaultRule(t *testing.T) {
	t.Parallel()

	list := newList([]Rule{
		{kind: pslNormal, value: "org", labels: 1},
		{kind: pslNormal, value: "om", labels: 1},
	}, nil, true)

	assert.Nil(t, list.Find("example.or"), "a suffix substring must not match")
	assert.Nil(t, list.Find("example.rg"))
	assert.Nil(t, list.Find("example.o"))
	assert.NotNil(t, list.Find("example.org"))
	assert.NotNil(t, list.Find("example.om"))
}

func TestList_FindNilReceiver(t *testing.T) {
	t.Parallel()

	var list *List

	assert.Nil(t, list.Find("example.com"))
}

func TestICANNExcludesPrivate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		list   *List
		host   string
		left   string
		suffix string
	}{
		{"default keeps a private suffix", DefaultList(), "blogspot.com", "", ""},
		{"icann skips a private suffix", ICANNList(), "blogspot.com", "blogspot", "com"},
		{"default under a private suffix", DefaultList(), "foo.blogspot.com", "foo", "blogspot.com"},
		{"icann under a private suffix", ICANNList(), "foo.blogspot.com", "foo.blogspot", "com"},
		{"icann skips a multi-label private suffix", ICANNList(), "s3.amazonaws.com", "s3.amazonaws", "com"},
		{"icann unknown tld uses the default rule", ICANNList(), "example.invalidtld", "example", "invalidtld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rule := tt.list.Find(tt.host)
			left, suffix := rule.Decompose(tt.host)
			assert.Equal(t, tt.left, left)
			assert.Equal(t, tt.suffix, suffix)
		})
	}
}
