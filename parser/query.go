package parser

import "strings"

// QueryPair is one raw key=value pair from a URL query string. Keys and values
// keep their original percent-encoding and ordering; unlike [url.Values], the
// representation is lossless for duplicate keys and repeated values.
//
// Fields:
//   - Key (string): the raw parameter name, percent-encoding preserved.
//   - Value (string): the raw parameter value, percent-encoding preserved;
//     empty when the pair carries no "=".
type QueryPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ParseQueryPairs splits a raw query string into pairs while preserving the
// original parameter order and duplicates. A pair without "=" is returned with
// an empty Value. Empty segments (for example the empty string between two
// "&&") are skipped.
//
// Parameters:
//   - raw (string): the raw query string, without the leading "?".
//
// Returns:
//   - pairs ([]QueryPair): the parsed pairs, in order of appearance.
func ParseQueryPairs(raw string) (pairs []QueryPair) {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, "&")

	pairs = make([]QueryPair, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			continue
		}

		key, value, _ := strings.Cut(part, "=")

		pairs = append(pairs, QueryPair{Key: key, Value: value})
	}

	return
}
