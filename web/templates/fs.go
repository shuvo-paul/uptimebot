package templates

import "embed"

//go:embed layouts/*.html
//go:embed pages/*.html
//go:embed pages/sites/*.html
var TemplateFS embed.FS
