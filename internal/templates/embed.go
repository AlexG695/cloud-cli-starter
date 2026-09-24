package templates

import "embed"

// FS contains all templates embedded in the compiled binary.
//
//go:embed *.tmpl
var FS embed.FS
