package templates

import "embed"

//go:embed layouts/*.html
//go:embed pages/*.html
//go:embed pages/targets/*.html
var TemplateFS embed.FS
