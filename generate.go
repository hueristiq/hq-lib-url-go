package url

// This file only holds the module's go:generate directives, which regenerate the data
// packages (PSL rules, schemes, tlds, unicodes) from their public sources. The package
// documentation lives in doc.go.

//go:generate sh -c "cd cmd/pslgen && go run . -output ../../internal/psl/rules_gen.go"
//go:generate sh -c "cd cmd/schemesgen && go run . -output ../../scheme/schemes_official_gen.go"
//go:generate sh -c "cd cmd/tldsgen && go run . -output ../../tld/tlds_official_gen.go"
//go:generate sh -c "cd cmd/unicodesgen && go run . -output ../../unicode/unicodes_gen.go"
