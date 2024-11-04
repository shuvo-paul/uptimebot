package views

import (
	"embed"
	"text/template"
)

type Template struct {
	fs embed.FS
}

func NewTemplate(fs embed.FS) *Template {
	return &Template{fs: fs}
}

func (t *Template) Parse(files ...string) *template.Template {
	// Convert *Template back to embed.FS and parse
	paths := append([]string{"layouts/base.html"}, files...)
	return template.Must(template.ParseFS(t.fs, paths...))
}
