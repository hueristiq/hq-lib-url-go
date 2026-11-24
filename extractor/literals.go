package extractor

import (
	"strings"
)

// ExpressionPlaceholder is substituted for an expression whose value cannot be known
// when a group of concatenated literals is collapsed.
const ExpressionPlaceholder = "EXPR"

// LiteralGroup is a run of adjacent string literals joined by the '+' operator in
// JavaScript source text. It covers plain quoted strings, template literals
// with ${...} interpolations, and unknown expressions between literals.
//
// Fields:
//   - Start (int): byte offset of the first literal's opening quote.
//   - End (int): byte offset just past the last literal's closing quote.
//   - Pieces ([]string): escape-decoded contents of the group's literals, in order;
//     unknown expressions contribute an empty piece.
//   - HasExpr (bool): whether an expression with an unknown value (or a ${...}
//     interpolation) participates in the group.
//   - exprJoins ([]bool): per piece after the first, whether the join preceding the
//     piece is an unknown expression rather than a literal-to-literal concatenation; a
//     missing entry is treated as a literal join.
type LiteralGroup struct {
	Start     int
	End       int
	Pieces    []string
	HasExpr   bool
	exprJoins []bool
}

// Collapsed returns the concatenation of the group's literal pieces with
// ExpressionPlaceholder standing in for every unknown expression.
func (g LiteralGroup) Collapsed() (collapsed string) {
	var out strings.Builder

	for i, piece := range g.Pieces {
		if i < len(g.exprJoins) && g.exprJoins[i] {
			out.WriteString(ExpressionPlaceholder)
		}

		out.WriteString(piece)
	}

	collapsed = out.String()

	return collapsed
}

// JSLiteralGroups scans text for string literals and returns the concatenation groups
// it finds, in source order. It skips line and block comments outside
// literals. A group always holds at least one piece; single-literal groups
// without an expression are returned too, leaving filtering to the caller.
//
// Literals inside an unknown expression are returned as separate groups. Those
// nested groups may overlap the Start/End range of the enclosing group.
//
// The scanner is text-based, not a parser: division expressions, regex
// literals, and unquoted markup may occasionally be read as literals.
func JSLiteralGroups(text string) (groups []LiteralGroup) {
	groups = make([]LiteralGroup, 0)

	for i := 0; i < len(text); {
		switch {
		case text[i] == '/' && i+1 < len(text) && text[i+1] == '/':
			i = skipLineComment(text, i+2)
		case text[i] == '/' && i+1 < len(text) && text[i+1] == '*':
			i = skipBlockComment(text, i+2)
		case isQuote(text[i]):
			group, nested, scanEnd := scanGroup(text, i)

			groups = append(groups, group)
			groups = append(groups, nested...)

			i = scanEnd
		default:
			i++
		}
	}

	return groups
}

// scanGroup scans the concatenation group starting with the literal opening
// at start.
func scanGroup(text string, start int) (group LiteralGroup, nested []LiteralGroup, scanEnd int) {
	pieces, hasExpr, joins, litEnd := scanLiteral(text, start)

	group = LiteralGroup{
		Start:     start,
		Pieces:    pieces,
		HasExpr:   hasExpr,
		exprJoins: joins,
	}

	lastLiteralEnd := litEnd
	pos := litEnd
	scanEnd = litEnd

	for {
		after := skipSpaceAndComments(text, pos)

		if after >= len(text) || text[after] != '+' {
			break
		}

		pos = skipSpaceAndComments(text, after+1)

		if pos >= len(text) {
			scanEnd = pos

			break
		}

		if isQuote(text[pos]) {
			literalPieces, literalExpr, literalJoins, literalEnd := scanLiteral(text, pos)

			group.Pieces = append(group.Pieces, literalPieces...)
			group.exprJoins = append(group.exprJoins, false)
			group.exprJoins = append(group.exprJoins, literalJoins[1:]...)
			group.HasExpr = group.HasExpr || literalExpr

			lastLiteralEnd = literalEnd
			pos = literalEnd
			scanEnd = literalEnd

			continue
		}

		// The operand is an expression whose value cannot be known. Record a
		// placeholder and skip to the next '+' at the same nesting level, or
		// to a statement boundary. Literals inside the expression are returned
		// as nested groups so their contents are not lost.
		group.Pieces = append(group.Pieces, "")
		group.exprJoins = append(group.exprJoins, true)
		group.HasExpr = true

		exprStart := pos

		pos = skipUnknown(text, pos)
		scanEnd = pos

		nested = append(nested, literalGroupsWithin(text, exprStart, pos)...)
	}

	group.End = lastLiteralEnd

	return group, nested, scanEnd
}

// literalGroupsWithin returns the literal groups found in text[start:end],
// with their offsets rebased to text. It recovers literals nested inside an
// unknown expression that scanGroup skips over.
func literalGroupsWithin(text string, start, end int) (groups []LiteralGroup) {
	if start >= end {
		return nil
	}

	groups = JSLiteralGroups(text[start:end])

	for i := range groups {
		groups[i].Start += start
		groups[i].End += start
	}

	return groups
}

// scanLiteral scans the literal whose opening quote is at start and returns
// its decoded pieces and the offset just past the literal's end.
func scanLiteral(text string, start int) (pieces []string, hasExpr bool, joins []bool, end int) {
	if text[start] == '`' {
		return scanTemplate(text, start)
	}

	pieces, joins, end = scanQuotedString(text, start, text[start])

	return pieces, false, joins, end
}

// scanQuotedString scans a '...' or "..." literal. Quoted strings cannot span
// lines in JavaScript; an unescaped line break ends the scan so malformed
// input does not swallow the rest of the document.
func scanQuotedString(text string, start int, quote byte) (pieces []string, joins []bool, end int) {
	pos := start + 1
	chunkStart := pos

	for pos < len(text) {
		switch text[pos] {
		case '\\':
			pos += escapeWidth(text, pos)
		case quote:
			pieces = []string{DecodeJSEscapes(text[chunkStart:pos])}
			joins = []bool{false}

			return pieces, joins, pos + 1
		case '\n', '\r':
			pieces = []string{DecodeJSEscapes(text[chunkStart:pos])}
			joins = []bool{false}

			return pieces, joins, pos
		default:
			pos++
		}
	}

	pieces = []string{DecodeJSEscapes(text[chunkStart:pos])}
	joins = []bool{false}

	return pieces, joins, pos
}

// scanTemplate scans a `...` template literal, splitting it at ${...}
// interpolations. Each interpolation sets hasExpr; its content is skipped.
func scanTemplate(text string, start int) (pieces []string, hasExpr bool, joins []bool, end int) {
	pos := start + 1
	chunkStart := pos
	exprJoin := false

	for {
		if pos >= len(text) {
			pieces = append(pieces, DecodeJSEscapes(text[chunkStart:pos]))
			joins = append(joins, exprJoin)

			return pieces, hasExpr, joins, pos
		}

		switch {
		case text[pos] == '`':
			pieces = append(pieces, DecodeJSEscapes(text[chunkStart:pos]))
			joins = append(joins, exprJoin)

			return pieces, hasExpr, joins, pos + 1
		case text[pos] == '\\':
			pos += escapeWidth(text, pos)
		case text[pos] == '$' && pos+1 < len(text) && text[pos+1] == '{':
			pieces = append(pieces, DecodeJSEscapes(text[chunkStart:pos]))
			joins = append(joins, exprJoin)
			hasExpr = true
			exprJoin = true

			pos = skipInterpolation(text, pos+2)
			chunkStart = pos
		default:
			pos++
		}
	}
}

// skipInterpolation skips the ${...} body starting just past the opening
// brace and returns the offset just past the matching '}'. Nested strings,
// templates, and comments are skipped so their braces do not count.
func skipInterpolation(text string, pos int) (next int) {
	depth := 1

	for pos < len(text) {
		switch {
		case isQuote(text[pos]):
			_, _, _, pos = scanLiteral(text, pos)
		case text[pos] == '/' && pos+1 < len(text) && text[pos+1] == '/':
			pos = skipLineComment(text, pos+2)
		case text[pos] == '/' && pos+1 < len(text) && text[pos+1] == '*':
			pos = skipBlockComment(text, pos+2)
		case text[pos] == '{':
			depth++
			pos++
		case text[pos] == '}':
			depth--

			if depth == 0 {
				return pos + 1
			}

			pos++
		default:
			pos++
		}
	}

	return pos
}

// skipUnknown skips the expression starting at pos and returns the offset of
// the next '+' at the same nesting level, or of a statement boundary
// (';', ',', a closing bracket, or a line break), or len(text).
func skipUnknown(text string, pos int) (next int) {
	depth := 0

	for pos < len(text) {
		switch {
		case isQuote(text[pos]):
			_, _, _, end := scanLiteral(text, pos)

			pos = end
		case text[pos] == '/' && pos+1 < len(text) && text[pos+1] == '/':
			pos = skipLineComment(text, pos+2)
		case text[pos] == '/' && pos+1 < len(text) && text[pos+1] == '*':
			pos = skipBlockComment(text, pos+2)
		case text[pos] == '(' || text[pos] == '[' || text[pos] == '{':
			depth++
			pos++
		case text[pos] == ')' || text[pos] == ']' || text[pos] == '}':
			if depth == 0 {
				return pos
			}

			depth--
			pos++
		case depth == 0 && (text[pos] == ';' || text[pos] == ',' || text[pos] == '\n' || text[pos] == '\r'):
			return pos
		case text[pos] == '+':
			if depth == 0 {
				return pos
			}

			pos++
		default:
			pos++
		}
	}

	return pos
}

// skipSpaceAndComments skips whitespace (including line breaks) and comments.
func skipSpaceAndComments(text string, pos int) (next int) {
	for pos < len(text) {
		switch {
		case isSpace(text[pos]):
			pos++
		case text[pos] == '/' && pos+1 < len(text) && text[pos+1] == '/':
			pos = skipLineComment(text, pos+2)
		case text[pos] == '/' && pos+1 < len(text) && text[pos+1] == '*':
			pos = skipBlockComment(text, pos+2)
		default:
			return pos
		}
	}

	return pos
}

// skipLineComment returns the offset of the line break that ends the comment
// starting at pos, or len(text) when the comment runs to the end.
func skipLineComment(text string, pos int) (next int) {
	for pos < len(text) && text[pos] != '\n' {
		pos++
	}

	return pos
}

// skipBlockComment returns the offset just past the block comment starting at
// pos, or len(text) when the comment is unterminated.
func skipBlockComment(text string, pos int) (next int) {
	for pos+1 < len(text) {
		if text[pos] == '*' && text[pos+1] == '/' {
			return pos + 2
		}

		pos++
	}

	return len(text)
}

// escapeWidth returns how many bytes the escape sequence starting at pos
// (text[pos] is the backslash) occupies, so the escaped character cannot be
// mistaken for a literal terminator.
func escapeWidth(text string, pos int) (width int) {
	if pos+1 >= len(text) {
		return 1
	}

	if text[pos+1] == '\r' && pos+2 < len(text) && text[pos+2] == '\n' {
		return 3
	}

	return 2
}

// isQuote reports whether b opens a JavaScript string literal.
func isQuote(b byte) (quote bool) {
	return b == '\'' || b == '"' || b == '`'
}

// isSpace reports whether b is JavaScript whitespace.
func isSpace(b byte) (space bool) {
	switch b {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}

	return false
}
