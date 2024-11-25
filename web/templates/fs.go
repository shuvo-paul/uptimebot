package templates

import "embed"

//go:embed layouts/*.html
//go:embed pages/*.html
var TemplateFS embed.FS
