package parser

import "strings"

// Domain represents a parsed domain name, broken down into three main components.
//
// Fields:
//   - TopLevelDomain (string): recognized top-level domain (e.g. "com" or "co.uk").
//   - SecondLevelDomain (string): label before the public suffix (e.g. "example" in
//     "www.example.com").
//   - Subdomain (string): labels before the second-level domain (e.g. "www").
type Domain struct {
	TopLevelDomain    string
	SecondLevelDomain string
	Subdomain         string
}

// String reconstructs a full domain name from its individual components. It joins the non-empty
// components using a dot ("."). Empty components are omitted from the final output.
//
// Examples:
//   - "www.example.com" when all components are provided.
//   - "example.com" when the subdomain is empty.
//   - "example" when only the SLD is provided.
//
// Returns:
//   - domain (string): the reconstructed domain.
func (d *Domain) String() (domain string) {
	parts := make([]string, 0, 3)

	if d.Subdomain != "" {
		parts = append(parts, d.Subdomain)
	}

	if d.SecondLevelDomain != "" {
		parts = append(parts, d.SecondLevelDomain)
	}

	if d.TopLevelDomain != "" {
		parts = append(parts, d.TopLevelDomain)
	}

	domain = strings.Join(parts, ".")

	return domain
}

// Apex returns the apex (registrable) domain: the second-level domain joined with the
// public suffix (e.g. "example.com" for "www.example.com"). When one of the two
// components is empty, Apex returns the other.
//
// Returns:
//   - apex (string): the apex domain.
func (d *Domain) Apex() (apex string) {
	switch {
	case d.SecondLevelDomain == "":
		apex = d.TopLevelDomain
	case d.TopLevelDomain == "":
		apex = d.SecondLevelDomain
	default:
		apex = d.SecondLevelDomain + "." + d.TopLevelDomain
	}

	return apex
}
