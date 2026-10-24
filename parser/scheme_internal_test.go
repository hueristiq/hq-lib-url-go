package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_hasNumericPort(t *testing.T) {
	t.Parallel()

	assert.True(t, hasNumericPort("example.com:8080"))
	assert.False(t, hasNumericPort("example.com"))
	assert.False(t, hasNumericPort("example.com:"))
	assert.False(t, hasNumericPort("example.com:80a"))
}
