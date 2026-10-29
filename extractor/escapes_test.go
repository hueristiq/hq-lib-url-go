package extractor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
)

func Test_DecodeJSEscapes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain text",
			in:   "no escapes here",
			want: "no escapes here",
		},
		{
			name: "empty",
			in:   "",
			want: "",
		},
		{
			name: "single character escapes",
			in:   `a\bb\fc\ned\rf\tg\vh\'i\"j\\k`,
			want: "a\bb\fc\ned\rf\tg\vh'i\"j\\k",
		},
		{
			name: "hex escapes",
			in:   `\x68ttps://example.com`,
			want: "https://example.com",
		},
		{
			name: "hex escape with escaped dot",
			in:   `example\x2ecom`,
			want: "example.com",
		},
		{
			name: "unicode escapes",
			in:   `\u0068\u0074\u0074\u0070\u0073://example.com`,
			want: "https://example.com",
		},
		{
			name: "braced code point escapes",
			in:   `\u{48}\u{65}\u{6c}\u{6c}\u{6f}`,
			want: "Hello",
		},
		{
			name: "json escaped slashes",
			in:   `https:\/\/example.com\/api\/v1`,
			want: "https://example.com/api/v1",
		},
		{
			name: "json body",
			in:   `{"url":"https:\/\/example.com\/x","amp":"\u0026"}`,
			want: `{"url":"https://example.com/x","amp":"&"}`,
		},
		{
			name: "octal escapes",
			in:   `\101\102\103`,
			want: "ABC",
		},
		{
			name: "octal escape equals sign",
			in:   `foo\075bar`,
			want: "foo=bar",
		},
		{
			name: "octal escape stops at non-octal digit",
			in:   `\08`,
			want: "\x008",
		},
		{
			name: "octal escape with leading four",
			in:   `\400`,
			want: " 0",
		},
		{
			name: "surrogate pair",
			in:   `\uD83D\uDE00`,
			want: "😀",
		},
		{
			name: "unpaired low surrogate",
			in:   `\uDE00`,
			want: "\uFFFD",
		},
		{
			name: "identity escapes",
			in:   `a\.b\dc`,
			want: "a.bdc",
		},
		{
			name: "line continuation",
			in:   "a\\\nb",
			want: "ab",
		},
		{
			name: "carriage return line continuation",
			in:   "a\\\r\nb",
			want: "ab",
		},
		{
			name: "trailing backslash",
			in:   `abc\`,
			want: `abc\`,
		},
		{
			name: "invalid hex escape decodes to the character",
			in:   `\xzz`,
			want: "xzz",
		},
		{
			name: "invalid unicode escape decodes to the character",
			in:   `\uzzzz`,
			want: "uzzzz",
		},
		{
			name: "malformed braced escape decodes to the character",
			in:   `\u{zz}`,
			want: "u{zz}",
		},
		{
			name: "unterminated braced escape decodes to the character",
			in:   `\u{`,
			want: "u{",
		},
		{
			name: "out of range braced escape is dropped",
			in:   `\u{110000}`,
			want: "",
		},
		{
			name: "overflowing braced escape is dropped",
			in:   `\u{10000000000000000}`,
			want: "",
		},
		{
			name: "overflowing braced escape with many digits is dropped",
			in:   `\u{FFFFFFFFFFFFFFFF1}`,
			want: "",
		},
		{
			name: "long braced escape with leading zeros",
			in:   `\u{000000000000000041}`,
			want: "A",
		},
		{
			name: "mixed raw body",
			in: "const u = \"https:\\/\\/example.com\";\n" +
				"const v = 'https://' + 'example.com';\n" +
				"// https://example.com/in-comment",
			want: "const u = \"https://example.com\";\n" +
				"const v = 'https://' + 'example.com';\n" +
				"// https://example.com/in-comment",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := hqgourlextractor.DecodeJSEscapes(testCase.in)

			assert.Equal(t, testCase.want, got)
		})
	}
}
