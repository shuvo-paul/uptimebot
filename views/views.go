package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"

	"github.com/shuvo-paul/sitemonitor/csrf"
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
	if len(files) == 0 {
		panic("template: no files provided to parse")
	}

	tpl := template.New("base.html")
	tpl.Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", fmt.Errorf("csrfField not implemented")
		},
		"currentUser": func() (*models.User, error) {
			return &models.User{}, fmt.Errorf("currentUser not implemented")
		},
	})
	paths := append([]string{"layouts/base.html"}, files[0])
	return template.Must(tpl.ParseFS(t.fs, paths...))
}

func (t *Template) Execute(w http.ResponseWriter, r *http.Request, tmpl *template.Template, data any) {
	tpl, err := tmpl.Clone()
	if err != nil {
		slog.Error("Cloning template", "Error", err)
		http.Error(w, "There was an error rendering the page", http.StatusInternalServerError)
	}

	tpl.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrf.GenerateCsrfField(r)
		},
		"currentUser": func() *models.User {
			user, _ := services.GetUser(r.Context())
			return user
		},
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var buf bytes.Buffer

	err = tpl.Execute(&buf, data)

	if err != nil {
		slog.Error("Executing template", "Error", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}

	io.Copy(w, &buf)
}
