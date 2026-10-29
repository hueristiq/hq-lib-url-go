package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func Test_InferScheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		host           string
		expectedScheme string
	}{
		{"example.com", "https"},
		{"example.com:8443", "https"},
		{"8.8.8.8", "https"},
		{"EXAMPLE.COM.", "https"},
		{"localhost", "http"},
		{"localhost:8080", "http"},
		{"foo.localhost", "http"},
		{"host.local", "http"},
		{"host.test", "http"},
		{"host.invalid", "http"},
		{"127.0.0.1", "http"},
		{"127.0.0.1:8080", "http"},
		{"::1", "http"},
		{"10.0.0.1", "http"},
		{"192.168.1.1", "http"},
		{"169.254.0.1", "http"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expectedScheme, hqgourlparser.InferScheme(tt.host))
		})
	}
}
