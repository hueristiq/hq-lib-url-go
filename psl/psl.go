package psl

import "strings"

// Public Suffix List rule kinds.
const (
	pslNormal uint8 = iota + 1
	pslWildcard
	pslException
)

// Rule is one rule from the embedded Public Suffix List.
//
// Fields:
//   - kind (uint8): pslNormal, pslWildcard, or pslException.
//   - value (string): the suffix without the "*." or "!" marker.
//   - labels (int): the number of labels the rule spans, including the wildcard label.
//   - private (bool): whether the rule belongs to the PSL private section.
type Rule struct {
	kind    uint8
	value   string
	labels  int
	private bool
}

// List is an immutable lookup table of Public Suffix List rules.
//
// Fields:
//   - rules (map[string]*Rule): rules keyed by their suffix value.
//   - defaultRule (*Rule): the implicit "*" rule, or nil when the list has none.
type List struct {
	rules       map[string]*Rule
	defaultRule *Rule
}

// defaultRule is the implicit "*" rule: when no explicit rule matches, the last
// label is the public suffix and the preceding label is the registrable label.
var defaultRule = &Rule{kind: pslWildcard, labels: 1}

// Default is the full embedded Public Suffix List generated from publicsuffix.org,
// including both the ICANN and private sections.
var Default = newList(pslRules[:], defaultRule, true)

// ICANN is the embedded Public Suffix List restricted to the ICANN section, excluding
// private-domain suffixes such as "blogspot.com" and "s3.amazonaws.com".
var ICANN = newList(pslRules[:], defaultRule, false)

// newList builds a lookup table from a set of rules.
//
// Parameters:
//   - rules ([]Rule): the rules to index.
//   - defaultRule (*Rule): the implicit rule to use when no explicit rule matches.
//   - includePrivate (bool): whether private-section rules are indexed.
//
// Returns:
//   - list (*List): the indexed rule set.
func newList(rules []Rule, defaultRule *Rule, includePrivate bool) (list *List) {
	list = &List{
		rules:       make(map[string]*Rule, len(rules)),
		defaultRule: defaultRule,
	}

	for i := range rules {
		rule := &rules[i]

		if !includePrivate && rule.private {
			continue
		}

		list.rules[rule.value] = rule
	}

	return list
}

// Find returns the most specific rule matching host, or the list's default rule when
// no explicit rule matches.
//
// Parameters:
//   - host (string): the lowercase hostname to look up.
//
// Returns:
//   - rule (*Rule): the matching rule, or nil when the list has no default rule.
func (l *List) Find(host string) (rule *Rule) {
	if l == nil {
		return nil
	}

	part := host

	for {
		if candidate, ok := l.rules[part]; ok && candidate.match(host) {
			return candidate
		}

		i := strings.IndexByte(part, '.')
		if i < 0 {
			return l.defaultRule
		}

		part = part[i+1:]
	}
}

// match reports whether r applies to host.
//
// Parameters:
//   - host (string): the lowercase hostname to match.
//
// Returns:
//   - matches (bool): whether the rule applies.
func (r *Rule) match(host string) (matches bool) {
	left := strings.TrimSuffix(host, r.value)

	if left == "" {
		return r.kind != pslWildcard
	}

	return strings.HasSuffix(left, ".")
}

// Decompose splits host into the labels left of the public suffix and the suffix itself.
//
// Parameters:
//   - host (string): the lowercase hostname to split.
//
// Returns:
//   - left (string): the labels before the public suffix.
//   - suffix (string): the public suffix.
func (r *Rule) Decompose(host string) (left, suffix string) {
	if r == defaultRule {
		i := strings.LastIndexByte(host, '.')
		if i < 0 {
			return "", ""
		}

		return host[:i], host[i+1:]
	}

	switch r.kind {
	case pslNormal:
		left = strings.TrimSuffix(host, r.value)
		if left == "" {
			return "", ""
		}

		left = left[:len(left)-1]
		suffix = r.value
	case pslWildcard:
		left = strings.TrimSuffix(host, r.value)
		if left == "" {
			return "", ""
		}

		left = left[:len(left)-1]

		i := strings.LastIndexByte(left, '.')
		if i < 0 {
			return "", ""
		}

		suffix = left[i+1:] + "." + r.value
		left = left[:i]
	case pslException:
		i := strings.IndexByte(r.value, '.')
		if i < 0 {
			return "", ""
		}

		suffix = r.value[i+1:]

		left = strings.TrimSuffix(host, suffix)
		if left == "" {
			return "", ""
		}

		left = left[:len(left)-1]
	}

	return left, suffix
}
