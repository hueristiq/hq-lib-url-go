#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

go run ./cmd/pslgen -output internal/psl/rules_gen.go
go run ./cmd/schemesgen -output internal/scheme/schemes_official_gen.go
go run ./cmd/tldsgen -output internal/tld/tlds_official_gen.go
go run ./cmd/unicodesgen -output internal/unicode/unicodes_gen.go
