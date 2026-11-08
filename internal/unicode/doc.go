// Package unicode holds the generated Unicode character ranges used to build
// the extractor's URL-matching expressions.
//
// # Entry points
//
//   - [AllowedUcsChar] is the generated regex character range of Unicode
//     characters allowed inside a URL component.
//   - [AllowedUcsCharMinusPunc] is the same range minus the punctuation a URL
//     component must not end with. Both are regenerated from the Unicode
//     character database by cmd/unicodesgen.
package unicode
