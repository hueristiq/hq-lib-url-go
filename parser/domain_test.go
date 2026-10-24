package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hueristiq/hq-lib-url-go/parser"
)

func Test_Domain_String_And_Apex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		domain       parser.Domain
		expectedFull string
		expectedApex string
	}{
		{
			"full domain",
			parser.Domain{Subdomain: "www", SecondLevelDomain: "example", TopLevelDomain: "com"},
			"www.example.com",
			"example.com",
		},
		{
			"without subdomain",
			parser.Domain{SecondLevelDomain: "example", TopLevelDomain: "com"},
			"example.com",
			"example.com",
		},
		{
			"single label",
			parser.Domain{SecondLevelDomain: "localhost"},
			"localhost",
			"localhost",
		},
		{
			"suffix only",
			parser.Domain{TopLevelDomain: "co.uk"},
			"co.uk",
			"co.uk",
		},
		{
			"zero domain",
			parser.Domain{},
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expectedFull, tt.domain.String())
			assert.Equal(t, tt.expectedApex, tt.domain.Apex())
		})
	}
}
