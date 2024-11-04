package templates

import "embed"

//go:embed layouts/*.html
//go:embed *.html
var TemplateFS embed.FS
