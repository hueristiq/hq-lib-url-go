package url

// This file only holds the module's go:generate directives, which regenerate the data
// packages (PSL rules, schemes, tlds, unicodes) from their public sources. The package
// documentation lives in doc.go.

//go:generate go run ./cmd/pslgen -output ./psl/rules.go
//go:generate go run ./cmd/schemesgen -output ./schemes/schemes_official.go
//go:generate go run ./cmd/tldsgen -output ./tlds/tlds_official.go
//go:generate go run ./cmd/unicodesgen -output ./unicodes/unicodes.go
