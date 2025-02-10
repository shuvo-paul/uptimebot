package templates

import "embed"

//go:embed layouts/*.html
//go:embed pages/*.html
//go:embed pages/targets/*.html
//go:embed emails/*.html
var TemplateFS embed.FS
