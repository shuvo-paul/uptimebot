package views

import (
	"embed"
	"net/http"
	"text/template"

	"github.com/shuvo-paul/sitemonitor/models"
	"github.com/shuvo-paul/sitemonitor/services"
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

func (t *Template) Execute(w http.ResponseWriter, r *http.Request, tmpl *template.Template, data any) error {
	tmpl.Funcs(template.FuncMap{
		"currentUser": func() (*models.User, bool) {
			return services.GetUser(r.Context())
		},
	})

	return tmpl.Execute(w, data)
}
