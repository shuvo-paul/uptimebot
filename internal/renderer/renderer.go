package renderer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/auth/service"
	"github.com/shuvo-paul/uptimebot/pkg/csrf"
)

type Engine struct {
	fs      embed.FS
	pages   map[string]*PageTemplate
	funcMap template.FuncMap
}

func New(fs embed.FS) *Engine {
	e := &Engine{
		fs:    fs,
		pages: make(map[string]*PageTemplate),
		funcMap: template.FuncMap{
			"csrfField": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"currentUser": func() (*model.User, error) {
				return &model.User{}, fmt.Errorf("currentUser not implemented")
			},
		},
	}

	pagesDir := "pages"

	entries, err := fs.ReadDir(pagesDir)
	if err != nil {
		slog.Error("error reading pages directory", "error", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && entry.Name()[len(entry.Name())-5:] == ".html" {
			path := pagesDir + "/" + entry.Name()
			name := entry.Name()[:5]
			fmt.Println(name)
			newTmpl := template.New(path)
			newTmpl.Funcs(e.funcMap)
			tmpl := template.Must(newTmpl.ParseFS(fs, path))
			e.pages[name] = &PageTemplate{tmpl: tmpl}
		}
	}

	return e
}

func (rndr *Engine) Parse(files string) PageTemplate {
	if files == "" {
		panic("template: no files provided to parse")
	}

	tpl := template.New("base.html").Funcs(rndr.funcMap)
	paths := append([]string{"layouts/base.html"}, "pages/"+files)
	tmpl := template.Must(tpl.ParseFS(rndr.fs, paths...))
	return PageTemplate{
		tmpl: tmpl,
	}
}

type PageTemplate struct {
	tmpl *template.Template
}

func (t *PageTemplate) Render(w http.ResponseWriter, r *http.Request, data any) {
	tpl, err := t.tmpl.Clone()
	if err != nil {
		slog.Error("cloning template", "error", err)
		http.Error(w, "There was an error rendering the page", http.StatusInternalServerError)
		return
	}

	tpl.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrf.GenerateCsrfField(r)
		},
		"currentUser": func() *model.User {
			user, _ := service.GetUser(r.Context())
			return user
		},
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		slog.Error("executing template", "error", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}

	io.Copy(w, &buf)
}
