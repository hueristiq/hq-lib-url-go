package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hqgourlparser "github.com/hueristiq/hq-lib-url-go/parser"
)

func Test_Domain_String_And_Apex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		domain       hqgourlparser.Domain
		expectedFull string
		expectedApex string
	}{
		{
			"full domain",
			hqgourlparser.Domain{Subdomain: "www", SecondLevelDomain: "example", TopLevelDomain: "com"},
			"www.example.com",
			"example.com",
		},
		{
			"without subdomain",
			hqgourlparser.Domain{SecondLevelDomain: "example", TopLevelDomain: "com"},
			"example.com",
			"example.com",
		},
		{
			"single label",
			hqgourlparser.Domain{SecondLevelDomain: "localhost"},
			"localhost",
			"localhost",
		},
		{
			"suffix only",
			hqgourlparser.Domain{TopLevelDomain: "co.uk"},
			"co.uk",
			"co.uk",
		},
		{
			"zero domain",
			hqgourlparser.Domain{},
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
