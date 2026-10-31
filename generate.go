package url

// This file only holds the module's go:generate directives, which regenerate the data
// packages (PSL rules, schemes, tlds, unicodes) from their public sources. The package
// documentation lives in doc.go.

//go:generate go run ./cmd/pslgen -output ./internal/psl/rules.go
//go:generate go run ./cmd/schemesgen -output ./internal/scheme/schemes_official.go
//go:generate go run ./cmd/tldsgen -output ./internal/tld/tlds_official.go
//go:generate go run ./cmd/unicodesgen -output ./internal/unicode/unicodes.go
