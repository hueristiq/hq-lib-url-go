package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func TestDecodeUnreservedEscapes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "plain", hqgourlparser.DecodeUnreservedEscapes("plain"))
	assert.Equal(t, "~user%2Fpath%2E%2E", hqgourlparser.DecodeUnreservedEscapes("%7Euser%2Fpath%2E%2E"))
	assert.Equal(t, "bad%zz", hqgourlparser.DecodeUnreservedEscapes("bad%zz"))
}
