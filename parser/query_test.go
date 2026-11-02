package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func TestParseQueryPairs(t *testing.T) {
	t.Parallel()

	assert.Empty(t, hqgourlparser.ParseQueryPairs(""))
	assert.Equal(t, []hqgourlparser.QueryPair{
		{Key: "a", Value: "1"},
		{Key: "b", Value: "2"},
		{Key: "a", Value: "3"},
		{Key: "flag", Value: ""},
		{Key: "empty", Value: ""},
	}, hqgourlparser.ParseQueryPairs("a=1&b=2&a=3&flag&empty=&"))
}
