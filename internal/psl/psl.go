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

var (
	// defaultList is the full embedded Public Suffix List, including both the
	// ICANN and private sections.
	defaultList = newList(pslRules[:], defaultRule, true)

	// icannList is the embedded Public Suffix List restricted to the ICANN
	// section, excluding private-domain suffixes such as "blogspot.com" and
	// "s3.amazonaws.com".
	icannList = newList(pslRules[:], defaultRule, false)
)

// DefaultList returns the full embedded Public Suffix List generated from
// publicsuffix.org, including both the ICANN and private sections. The
// returned list is immutable and safe for concurrent use.
//
// Returns:
//   - list (*List): the full embedded list.
func DefaultList() (list *List) {
	return defaultList
}

// ICANNList returns the embedded Public Suffix List restricted to the ICANN
// section, excluding private-domain suffixes such as "blogspot.com" and
// "s3.amazonaws.com". The returned list is immutable and safe for concurrent
// use.
//
// Returns:
//   - list (*List): the ICANN-only list.
func ICANNList() (list *List) {
	return icannList
}

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

// Kind returns the rule kind: "normal", "wildcard", or "exception". It
// returns "" for a nil rule.
//
// Returns:
//   - kind (string): the rule kind.
func (r *Rule) Kind() (kind string) {
	if r == nil {
		return ""
	}

	switch r.kind {
	case pslNormal:
		return "normal"
	case pslWildcard:
		return "wildcard"
	case pslException:
		return "exception"
	default:
		return ""
	}
}

// Suffix returns the rule suffix without the PSL wildcard or exception
// marker. It returns "" for a nil rule.
//
// Returns:
//   - suffix (string): the rule suffix.
func (r *Rule) Suffix() (suffix string) {
	if r == nil {
		return ""
	}

	return r.value
}

// Private reports whether the rule belongs to the PSL private section. It
// returns false for a nil rule.
//
// Returns:
//   - private (bool): whether the rule is a private-section rule.
func (r *Rule) Private() (private bool) {
	if r == nil {
		return false
	}

	return r.private
}

// String returns the canonical PSL text for the rule: the value prefixed with
// "!" for an exception or "*." for a wildcard. It returns "" for a nil rule.
//
// Returns:
//   - rule (string): the canonical rule text.
func (r *Rule) String() (rule string) {
	if r == nil {
		return ""
	}

	switch r.kind {
	case pslWildcard:
		return "*." + r.value
	case pslException:
		return "!" + r.value
	default:
		return r.value
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
