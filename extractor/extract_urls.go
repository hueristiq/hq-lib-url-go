package extractor

import (
	"errors"
	"fmt"
	"html"
	"net/url"
	"path"
	"regexp"
	"strings"
)

// ExtractURLs is the one-call form of [Extractor.ExtractURLs] using the default
// extractor configured by ofs.
//
// Parameters:
//   - text (string): the text to search.
//   - ofs (...OptionFunc): extractor configuration options.
//
// Returns:
//   - urls ([]string): the cleaned URLs found, in order of appearance.
//   - err (error): non-nil when the extraction pattern cannot be compiled.
func ExtractURLs(text string, ofs ...OptionFunc) (urls []string, err error) {
	extractor, err := New(ofs...)
	if err != nil {
		return nil, err
	}

	return extractor.ExtractURLs(text), nil
}

// ExtractURLs finds every URL in text and applies the boundary and cleanup
// rules a scanner needs: a match must not start or end inside a longer host
// label, a "%2F" prefix left behind by a percent-encoded slash is trimmed, and
// rows glued together by a source-code delimiter are split into individual
// URLs.
//
// The extraction pattern is compiled once by [New].
//
// Parameters:
//   - text (string): the text to search.
//
// Returns:
//   - urls ([]string): the cleaned URLs found, in order of appearance.
func (e *Extractor) ExtractURLs(text string) (urls []string) {
	for _, indexes := range e.regexp.FindAllStringIndex(text, -1) {
		start := indexes[0]

		if start > 0 && isHostLabelCharacter(text[start-1]) {
			continue
		}

		end := indexes[1]

		if end < len(text) && isHostLabelCharacter(text[end]) {
			if text[end] != '.' || end+1 < len(text) && isHostLabelCharacter(text[end+1]) {
				continue
			}
		}

		match, ok := trimEncodedSlashPrefix(text, start, text[start:end])
		if !ok {
			continue
		}

		urls = append(urls, splitGluedURLs(match)...)
	}

	return
}

// ExtractURLsAll extracts URLs from rawText and from its HTML- and
// JavaScript-unescaped form, deduplicating the combined result. It is the
// scanner entry point for JSON, JavaScript, and HTML bodies that may carry
// escaped URLs.
//
// Parameters:
//   - text (string): the text to search.
//   - ofs (...OptionFunc): extractor configuration options.
//
// Returns:
//   - urls ([]string): the unique URLs found, in order of first appearance.
//   - err (error): non-nil when the extraction pattern cannot be compiled.
func ExtractURLsAll(text string, ofs ...OptionFunc) (urls []string, err error) {
	extractor, err := New(ofs...)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})

	collect := func(value string) {
		for _, match := range extractor.ExtractURLs(value) {
			if _, exists := seen[match]; exists {
				continue
			}

			seen[match] = struct{}{}

			urls = append(urls, match)
		}
	}

	collect(text)

	if decoded := DecodeJSEscapes(html.UnescapeString(text)); decoded != text {
		collect(decoded)
	}

	return
}

// ExtractCollapsedURLs extracts URLs from JavaScript string-literal groups
// whose concatenations or template interpolations were collapsed by
// [JSLiteralGroups].
//
// Parameters:
//   - text (string): the JavaScript text to scan.
//   - ofs (...OptionFunc): extractor configuration options.
//
// Returns:
//   - urls ([]string): the URLs found in the collapsed literals.
//   - err (error): non-nil when the extraction pattern cannot be compiled.
func ExtractCollapsedURLs(text string, ofs ...OptionFunc) (urls []string, err error) {
	extractor, err := New(ofs...)
	if err != nil {
		return nil, err
	}

	for _, group := range JSLiteralGroups(text) {
		if len(group.Pieces) == 1 && !group.HasExpr {
			continue
		}

		collapsed := group.Collapsed()

		if strings.Trim(collapsed, ExpressionPlaceholder+"/") == "" {
			continue
		}

		urls = append(urls, extractor.ExtractURLs(collapsed)...)
	}

	return
}

// ExtractSubdomains finds hosts under domain in text, preserving the match
// order and applying the same boundary checks as [Extractor.ExtractURLs]. The
// domain is matched literally; call [NormalizeHost] first when the input may be
// a Unicode or mixed-case name.
//
// Parameters:
//   - text (string): the text to search.
//   - domain (string): the parent domain to match.
//
// Returns:
//   - subdomains ([]string): the matching hosts.
//   - err (error): non-nil when domain is empty or the pattern cannot compile.
func ExtractSubdomains(text, domain string) (subdomains []string, err error) {
	if domain == "" {
		err = errors.New("extractor.ExtractSubdomains: domain is empty")

		return
	}

	regex, err := regexp.Compile(fmt.Sprintf(
		`(?i)(?:^|[^a-z0-9_.-])(((?:[a-z0-9_](?:[a-z0-9_-]*[a-z0-9_])?\.)+)%s)`,
		regexp.QuoteMeta(domain),
	))
	if err != nil {
		err = fmt.Errorf("extractor.ExtractSubdomains: %w", err)

		return
	}

	for _, match := range regex.FindAllStringSubmatchIndex(text, -1) {
		start, end := match[2], match[3]

		value, ok := trimEncodedSlashPrefix(text, start, text[start:end])
		if !ok {
			continue
		}

		if end < len(text) && isHostLabelCharacter(text[end]) {
			if text[end] != '.' || end+1 < len(text) && isHostLabelCharacter(text[end+1]) {
				continue
			}
		}

		subdomains = append(subdomains, value)
	}

	return
}

// splitGluedURLs splits a match that contains several URLs joined by a
// source-code delimiter (",", ";", or "|") into individual URLs.
//
// Parameters:
//   - match (string): the raw match.
//
// Returns:
//   - urls ([]string): the individual URLs.
func splitGluedURLs(match string) (urls []string) {
	glue := gluedSchemeIndexes(match)

	if len(glue) == 0 {
		if value := trimGluedColumns(match); value != "" {
			return []string{value}
		}

		return nil
	}

	urls = make([]string, 0, len(glue)+1)

	start := 0

	for _, index := range glue {
		value := match[start:index]
		start = index + 1

		if value = trimGluedColumns(value); value != "" {
			urls = append(urls, value)
		}
	}

	if value := trimGluedColumns(match[start:]); value != "" {
		urls = append(urls, value)
	}

	return
}

// trimGluedColumns cuts source-code syntax that the path pattern greedily
// includes after a URL.
//
// Parameters:
//   - value (string): the candidate URL.
//
// Returns:
//   - trimmed (string): the cleaned URL.
func trimGluedColumns(value string) (trimmed string) {
	trimmed = trimCodeGlue(value)

	if status := gluedStatusIndex(trimmed); status >= 0 {
		return trimmed[:status]
	}

	if method := gluedMethodIndex(trimmed); method >= 0 {
		trimmed = trimmed[:method]
	}

	return
}

// trimCodeGlue cuts quotes, commas, and trailing source-code columns from a
// candidate URL.
//
// Parameters:
//   - value (string): the candidate URL.
//
// Returns:
//   - trimmed (string): the cleaned URL.
func trimCodeGlue(value string) (trimmed string) {
	trimmed = value

	if index := firstQuoteComma(trimmed); index >= 0 {
		trimmed = trimmed[:index]
	}

	index := strings.LastIndexByte(trimmed, ',')
	if index <= 0 {
		return trimmed
	}

	column := trimmed[index+1:]

	if !isSimpleColumn(column) || !endsLikeURL(trimmed[:index]) {
		return trimmed
	}

	trimmed = trimmed[:index]

	return
}

// firstQuoteComma returns the index of a quote adjacent to a comma, or -1.
//
// Parameters:
//   - value (string): the text to inspect.
//
// Returns:
//   - index (int): the position, or -1 when absent.
func firstQuoteComma(value string) (index int) {
	index = -1

	for position := 0; position+1 < len(value); position++ {
		first := value[position]
		second := value[position+1]

		if ((first == '\'' || first == '"') && second == ',') || (first == ',' && (second == '\'' || second == '"')) {
			return position
		}
	}

	return
}

// isSimpleColumn reports whether column looks like a source-code identifier or
// numeric column rather than URL path text.
//
// Parameters:
//   - column (string): the trailing column.
//
// Returns:
//   - simple (bool): true for letters, digits, "_", and "-".
func isSimpleColumn(column string) (simple bool) {
	if column == "" {
		return false
	}

	for _, character := range column {
		switch {
		case character >= 'a' && character <= 'z':
		case character >= 'A' && character <= 'Z':
		case character >= '0' && character <= '9':
		case character == '_', character == '-':
		default:
			return false
		}
	}

	return true
}

// endsLikeURL reports whether value ends like a complete URL path.
//
// Parameters:
//   - value (string): the candidate URL.
//
// Returns:
//   - ends (bool): true when the path ends with "/" or has an extension.
func endsLikeURL(value string) (ends bool) {
	if strings.HasSuffix(value, "/") {
		return true
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	return path.Ext(parsed.Path) != ""
}

// trimEncodedSlashPrefix removes the "2f" left behind when a match starts at
// the percent-encoded slash in "%2F". The match must be directly preceded by
// '%'; a remainder that is empty or starts with a dot is not a recoverable
// hostname and is dropped.
//
// Parameters:
//   - text (string): the full text.
//   - start (int): the match start offset in text.
//   - value (string): the raw match.
//
// Returns:
//   - trimmed (string): the cleaned match.
//   - ok (bool): false when the match is not a recoverable hostname.
func trimEncodedSlashPrefix(text string, start int, value string) (trimmed string, ok bool) {
	if start == 0 || text[start-1] != '%' || len(value) < 3 || !strings.EqualFold(value[:2], "2f") {
		return value, true
	}

	trimmed = value[2:]

	if trimmed == "" || trimmed[0] == '.' {
		return "", false
	}

	return trimmed, true
}

// isHostLabelCharacter reports whether character can continue a host label.
//
// Parameters:
//   - character (byte): the byte to inspect.
//
// Returns:
//   - ok (bool): true for letters, digits, ".", "-", and "_".
func isHostLabelCharacter(character byte) (ok bool) {
	switch {
	case character >= 'a' && character <= 'z', character >= 'A' && character <= 'Z', character >= '0' && character <= '9', character == '.', character == '-', character == '_':
		ok = true
	}

	return
}

// isGlueDelimiter reports whether character can glue two URLs together in
// source text.
//
// Parameters:
//   - character (byte): the byte to inspect.
//
// Returns:
//   - ok (bool): true for ",", ";", and "|".
func isGlueDelimiter(character byte) (ok bool) {
	return character == ',' || character == ';' || character == '|'
}

// gluedSchemeIndexes returns the offsets of the delimiters in match that are
// directly followed by an "http://" or "https://" scheme — the points where a
// new URL begins inside a glued row.
//
// Parameters:
//   - match (string): the raw match to scan.
//
// Returns:
//   - indexes ([]int): the delimiter offsets, in ascending order; nil when the
//     match holds a single URL.
func gluedSchemeIndexes(match string) (indexes []int) {
	for i := 0; i < len(match); i++ {
		if !isGlueDelimiter(match[i]) {
			continue
		}

		length := 0

		switch {
		case len(match)-i >= len(",https://") && strings.EqualFold(match[i:i+len(",https://")], ",https://"):
			length = len(",https://")
		case len(match)-i >= len(",http://") && strings.EqualFold(match[i:i+len(",http://")], ",http://"):
			length = len(",http://")
		}

		if length > 0 {
			indexes = append(indexes, i)
			i += length - 1
		}
	}

	return indexes
}

// gluedStatusIndex returns the offset of the last delimiter in match that is
// followed by a three-digit HTTP status code (and another delimiter or the end
// of the match) — the point where an access-log row stops being a URL.
//
// Parameters:
//   - match (string): the raw match to scan.
//
// Returns:
//   - index (int): the delimiter offset, or -1 when no status column is glued
//     on.
func gluedStatusIndex(match string) (index int) {
	index = -1

	for i := 0; i+4 < len(match); i++ {
		if !isGlueDelimiter(match[i]) {
			continue
		}

		if match[i+1] < '0' || match[i+1] > '9' {
			continue
		}

		if match[i+2] < '0' || match[i+2] > '9' {
			continue
		}

		if match[i+3] < '0' || match[i+3] > '9' {
			continue
		}

		if i+5 < len(match) && !isGlueDelimiter(match[i+4]) {
			continue
		}

		index = i
	}

	return index
}

// gluedMethodIndex returns the offset of the last delimiter in match that is
// followed by what looks like an HTTP method — two to eight uppercase letters —
// the point where an access-log row stops being a URL.
//
// Parameters:
//   - match (string): the raw match to scan.
//
// Returns:
//   - index (int): the delimiter offset, or -1 when no method column is glued
//     on.
func gluedMethodIndex(match string) (index int) {
	index = strings.LastIndexAny(match, ",;|")
	if index < 0 {
		return -1
	}

	method := match[index+1:]
	if len(method) < 2 || len(method) > 8 {
		return -1
	}

	for _, character := range method {
		if character < 'A' || character > 'Z' {
			return -1
		}
	}

	return index
}
