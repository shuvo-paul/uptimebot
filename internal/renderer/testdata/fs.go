package testdata

import "embed"

//go:embed layouts/*.html
//go:embed pages/*.html
var FS embed.FS
