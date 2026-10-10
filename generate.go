// This file only holds the module's go:generate directives, which regenerate the data
// packages (schemes, tlds, unicodes) from their public sources. The package
// documentation lives in doc.go.
package url

//go:generate go run schemes/gen/main.go -output ./schemes/schemes_official.go
//go:generate go run tlds/gen/main.go -output ./tlds/tlds_official.go
//go:generate go run unicodes/gen/main.go -output ./unicodes/unicodes.go
