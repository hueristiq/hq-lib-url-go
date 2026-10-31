//nolint:testpackage // tests exercise unexported behavior.
package psl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleAccessors(t *testing.T) {
	t.Parallel()

	normal := DefaultList().Find("example.co.uk")
	require.NotNil(t, normal)
	assert.Equal(t, "normal", normal.Kind())
	assert.Equal(t, "co.uk", normal.Suffix())
	assert.Equal(t, "co.uk", normal.String())
	assert.False(t, normal.Private())

	wildcard := DefaultList().Find("foo.ck")
	require.NotNil(t, wildcard)
	assert.Equal(t, "wildcard", wildcard.Kind())
	assert.Equal(t, "*.ck", wildcard.String())

	exception := DefaultList().Find("www.ck")
	require.NotNil(t, exception)
	assert.Equal(t, "exception", exception.Kind())
	assert.Equal(t, "!www.ck", exception.String())

	var nilRule *Rule

	assert.Empty(t, nilRule.Kind())
	assert.Empty(t, nilRule.Suffix())
	assert.Empty(t, nilRule.String())
	assert.False(t, nilRule.Private())
}
