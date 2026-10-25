package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hueristiq/hq-lib-url-go/parser"
)

func TestRepairInvalidEscapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "valid escapes",
			raw:  "https://example.com/a%2Fb/%0d%2e/",
			want: "https://example.com/a%2Fb/%0d%2e/",
		},
		{
			name: "invalid escape",
			raw:  "https://example.com/%invalid",
			want: "https://example.com/%25invalid",
		},
		{
			name: "trailing percent",
			raw:  "https://example.com/%",
			want: "https://example.com/%25",
		},
		{
			name: "short escape",
			raw:  "https://example.com/%A",
			want: "https://example.com/%25A",
		},
		{
			name: "non-hex escape",
			raw:  "https://example.com/%ZZ",
			want: "https://example.com/%25ZZ",
		},
		{
			name: "mixed path and query",
			raw:  "https://example.com/a%2Fb/%invalid?q=%zz&ok=%2F",
			want: "https://example.com/a%2Fb/%25invalid?q=%25zz&ok=%2F",
		},
		{
			name: "no percent",
			raw:  "https://example.com/path?q=1",
			want: "https://example.com/path?q=1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, parser.RepairInvalidEscapes(test.raw))
		})
	}
}
