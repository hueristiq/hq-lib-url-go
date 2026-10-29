package extractor_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
)

func Test_JSLiteralGroups(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   string
		want []string // Collapsed() of each group
	}{
		{
			name: "single literal",
			in:   `const a = "hello";`,
			want: []string{"hello"},
		},
		{
			name: "two groups",
			in:   `var a = "x"; var b = "y";`,
			want: []string{"x", "y"},
		},
		{
			name: "adjacent literals without plus do not merge",
			in:   `call("a", "b");`,
			want: []string{"a", "b"},
		},
		{
			name: "concatenated literals merge",
			in:   `var u = "https://" + "example.com" + "/api";`,
			want: []string{"https://example.com/api"},
		},
		{
			name: "expression between literals",
			in:   `var u = "https://" + host + "/api";`,
			want: []string{"https://EXPR/api"},
		},
		{
			name: "leading and trailing expressions",
			in:   `var u = proto + "://example.com" + path;`,
			want: []string{"://example.comEXPR"},
		},
		{
			name: "string inside skipped expression is returned separately",
			in:   `var u = "a" + fn("x") + "b";`,
			want: []string{"aEXPRb", "x"},
		},
		{
			name: "multiline concatenation",
			in:   "var u = \"a\" +\n\t\"b\";",
			want: []string{"ab"},
		},
		{
			name: "comment between operands",
			in:   `var u = "a" /* join */ + "b";`,
			want: []string{"ab"},
		},
		{
			name: "line comment between operands",
			in:   "var u = \"a\" + // join\n \"b\";",
			want: []string{"ab"},
		},
		{
			name: "template literal",
			in:   "var u = `https://example.com/api`;",
			want: []string{"https://example.com/api"},
		},
		{
			name: "template literal with interpolation",
			in:   "var u = `https://${host}/api`;",
			want: []string{"https://EXPR/api"},
		},
		{
			name: "template literal with nested template in interpolation",
			in:   "var u = `a${`nested ${x}`}b`;",
			want: []string{"aEXPRb"},
		},
		{
			name: "template literal joined by plus",
			in:   "var u = `https://` + host + `/api`;",
			want: []string{"https://EXPR/api"},
		},
		{
			name: "all expression template",
			in:   "var u = `${x}`;",
			want: []string{"EXPR"},
		},
		{
			name: "escaped quotes do not terminate literals",
			in:   `var u = "a\"b" + 'c' + "d";`,
			want: []string{`a"bcd`},
		},
		{
			name: "escaped unicode does not terminate literals",
			in:   `var u = "\u0022x" + "y";`,
			want: []string{`"xy`},
		},
		{
			name: "line continuation inside literal",
			in:   "var u = \"a\\\nb\" + \"c\";",
			want: []string{"abc"},
		},
		{
			name: "unterminated literal stops at line break",
			in:   "var u = \"abc\nvar v = \"def\";",
			want: []string{"abc", "def"},
		},
		{
			name: "unterminated literal at end of input",
			in:   `var u = "abc`,
			want: []string{"abc"},
		},
		{
			name: "comments are skipped",
			in:   "// \"not a literal\"\n/* \"nope\" */\nvar u = \"yes\";",
			want: []string{"yes"},
		},
		{
			name: "call chain in unknown expression",
			in:   `var u = "a" + b(c, "d") + "e";`,
			want: []string{"aEXPRe", "d"},
		},
		{
			name: "unknown expression with plus inside",
			in:   `var u = "a" + (b + c) + "d";`,
			want: []string{"aEXPRd"},
		},
		{
			name: "object literal in unknown expression",
			in:   `var u = "a" + {x: "y"} + "b";`,
			want: []string{"aEXPRb", "y"},
		},
		{
			name: "trailing expression literal is returned once",
			in:   `var u = "a" + fn("url");`,
			want: []string{"aEXPR", "url"},
		},
		{
			name: "nested expression literal groups",
			in:   `var u = "a" + fn("b" + g("c")) + "d";`,
			want: []string{"aEXPRd", "bEXPR", "c"},
		},
		{
			name: "division does not merge literals",
			in:   `var u = a / "b" / "c";`,
			want: []string{"b", "c"},
		},
		{
			name: "fetch with relative concatenation",
			in:   `fetch("/api/" + id + "/items");`,
			want: []string{"/api/EXPR/items"},
		},
		{
			name: "html attribute value scanned as one literal",
			in:   `onclick="fetch('/api/' + id)"`,
			want: []string{"fetch('/api/' + id)"},
		},
		{
			name: "empty input",
			in:   "",
			want: []string{},
		},
		{
			name: "no literals",
			in:   "var u = a + b;",
			want: []string{},
		},
		{
			name: "trailing plus with no operand",
			in:   `var u = "a" +`,
			want: []string{"a"},
		},
		{
			name: "unterminated template literal at end of input",
			in:   "var u = `abc",
			want: []string{"abc"},
		},
		{
			name: "escaped character in template literal",
			in:   "var u = `a\\nb`;",
			want: []string{"a\nb"},
		},
		{
			name: "line comment inside interpolation",
			in:   "var u = `a${x // c\n}b`;",
			want: []string{"aEXPRb"},
		},
		{
			name: "block comment inside interpolation",
			in:   "var u = `a${x /* } */}b`;",
			want: []string{"aEXPRb"},
		},
		{
			name: "nested braces inside interpolation",
			in:   "var u = `a${ {x: 1} }b`;",
			want: []string{"aEXPRb"},
		},
		{
			name: "unterminated interpolation",
			in:   "var u = `a${x",
			want: []string{"aEXPR"},
		},
		{
			name: "line comment inside unknown expression",
			in:   "var u = \"a\" + x // c\n + \"b\";",
			want: []string{"aEXPRb"},
		},
		{
			name: "block comment inside unknown expression",
			in:   `var u = "a" + x /* c */ + "b";`,
			want: []string{"aEXPRb"},
		},
		{
			name: "unmatched closing bracket in unknown expression",
			in:   `var u = "a" + ) + "b";`,
			want: []string{"aEXPR", "b"},
		},
		{
			name: "unknown expression runs to end of input",
			in:   `var u = "a" + host`,
			want: []string{"aEXPR"},
		},
		{
			name: "unterminated block comment after literal",
			in:   `var u = "a" /* unterminated`,
			want: []string{"a"},
		},
		{
			name: "trailing backslash in literal",
			in:   "var u = \"abc\\",
			want: []string{"abc\\"},
		},
		{
			name: "escaped CRLF inside literal",
			in:   "var u = \"a\\\r\nb\";",
			want: []string{"ab"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			groups := hqgourlextractor.JSLiteralGroups(testCase.in)

			collapsed := make([]string, 0, len(groups))

			for _, group := range groups {
				collapsed = append(collapsed, group.Collapsed())
			}

			assert.Equal(t, testCase.want, collapsed)
		})
	}
}

func Test_Groups_flags(t *testing.T) {
	t.Parallel()

	groups := hqgourlextractor.JSLiteralGroups(`var u = "https://" + host + "/api"; var v = "plain";`)

	require.Len(t, groups, 2)

	assert.True(t, groups[0].HasExpr)
	assert.Equal(t, []string{"https://", "", "/api"}, groups[0].Pieces)
	assert.Equal(t, "https://EXPR/api", groups[0].Collapsed())

	assert.False(t, groups[1].HasExpr)
	assert.Equal(t, []string{"plain"}, groups[1].Pieces)
}

func Test_Groups_offsets(t *testing.T) {
	t.Parallel()

	groups := hqgourlextractor.JSLiteralGroups(`var u = "abc" + "def";`)

	require.Len(t, groups, 1)

	assert.Equal(t, 8, groups[0].Start)
	assert.Equal(t, 21, groups[0].End)
}

func Test_LiteralGroup_Collapsed_externalConstruction(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "ab", hqgourlextractor.LiteralGroup{Pieces: []string{"a", "b"}}.Collapsed())
	assert.Empty(t, hqgourlextractor.LiteralGroup{}.Collapsed())
}

func Test_Groups_nestedOffsets(t *testing.T) {
	t.Parallel()

	text := `var u = "a" + fn("url") + "b";`

	groups := hqgourlextractor.JSLiteralGroups(text)

	require.Len(t, groups, 2)
	assert.Equal(t, "aEXPRb", groups[0].Collapsed())
	assert.Equal(t, "url", groups[1].Collapsed())

	start := strings.Index(text, `"url"`)

	assert.Equal(t, start, groups[1].Start)
	assert.Equal(t, start+len(`"url"`), groups[1].End)

	for i := 1; i < len(groups); i++ {
		assert.LessOrEqual(t, groups[i-1].Start, groups[i].Start)
	}
}
