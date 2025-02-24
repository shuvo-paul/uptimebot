package renderer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/auth/service"
	"github.com/shuvo-paul/uptimebot/pkg/csrf"
	"github.com/shuvo-paul/uptimebot/pkg/flash"
)

const layoutDir = "layouts"

type Engine struct {
	fs         embed.FS
	templates  map[string]*Template
	layouts    []string // List of layout files
	flashStore flash.FlashStoreInterface
}

func New(fs embed.FS, flashStore flash.FlashStoreInterface) *Engine {
	e := &Engine{
		fs:         fs,
		templates:  make(map[string]*Template),
		layouts:    make([]string, 0),
		flashStore: flashStore,
	}

	if err := e.parseAllTemplates(); err != nil {
		slog.Error("failed to parse templates", "error", err)
	}

	return e
}

func (e *Engine) parseAllTemplates() error {
	// Discover all templates recursively
	if err := e.discoverTemplates("."); err != nil {
		return fmt.Errorf("failed to discover templates: %w", err)
	}

	if len(e.layouts) == 0 {
		return fmt.Errorf("no layout templates found in %s", layoutDir)
	}

	return nil
}

// discoverTemplates recursively finds all template files
// For layouts directory, it collects layout files
// For other directories, it parses the templates
func (e *Engine) discoverTemplates(dir string) error {
	entries, err := e.fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		fullPath := path.Join(dir, entry.Name())

		if entry.IsDir() {
			// Special handling for layouts directory
			if entry.Name() == layoutDir {
				layouts, err := e.findLayoutFiles(fullPath)
				if err != nil {
					return err
				}
				e.layouts = layouts
				continue
			}

			// Recursively discover templates in other directories
			if err := e.discoverTemplates(fullPath); err != nil {
				return err
			}
			continue
		}

		// Skip non-html files and files in root directory
		if !strings.HasSuffix(entry.Name(), ".html") || dir == "." {
			continue
		}

		// Skip layout files as they are handled separately
		if strings.HasPrefix(dir, layoutDir) {
			continue
		}

		if err := e.parseTemplate(dir, fullPath); err != nil {
			return err
		}
	}

	return nil
}

// findLayoutFiles finds all HTML files in the layouts directory
func (e *Engine) findLayoutFiles(dir string) ([]string, error) {
	entries, err := e.fs.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read layouts directory: %w", err)
	}

	var layouts []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			layouts = append(layouts, path.Join(dir, entry.Name()))
		}
	}

	return layouts, nil
}

// parseTemplate creates and parses a single template with its layouts
func (e *Engine) parseTemplate(dir, fullPath string) error {
	templateName := path.Base(fullPath)
	prefix := strings.Split(dir, "/")[0] // Get root directory (emails or pages)
	relativePath := strings.TrimPrefix(fullPath, prefix+"/")
	relativePath = strings.TrimSuffix(relativePath, ".html")
	key := fmt.Sprintf("%s:%s", prefix, relativePath)

	// Create a new template set with the filename as name
	tmpl := template.New(templateName)
	// Add template functions for all templates
	tmpl = tmpl.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return "" // This will be replaced at render time
		},
		"currentUser": func() *model.User {
			return nil // This will be replaced at render time
		},
		"flashMessages": func() (map[string][]string, error) {
			return map[string][]string{
				"successes": {},
				"errors":    {},
			}, nil // This will be replaced at render time
		},
	})

	pattern := []string{fullPath}
	pattern = append(pattern, e.layouts...)

	// Parse the content template and layouts
	tmpl, err := tmpl.ParseFS(e.fs, pattern...)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", key, err)
	}

	// Use the NewTemplate constructor
	e.templates[key] = NewTemplate(tmpl, e.flashStore)
	return nil
}

// GetTemplate returns a template by its full key
// Examples:
//   - GetTemplate("pages:index")
//   - GetTemplate("pages:targets/list")
//   - GetTemplate("emails:verify_email")
func (e *Engine) GetTemplate(key string) *Template {
	tmpl, ok := e.templates[key]
	if !ok {
		slog.Error("template not found", "key", key)
		panic(fmt.Sprintf("template not found: %s", key))
	}
	return tmpl
}

type Template struct {
	tmpl       *template.Template
	flashStore flash.FlashStoreInterface
}

// NewTemplate creates a new Template instance
func NewTemplate(tmpl *template.Template, flashStore flash.FlashStoreInterface) *Template {
	return &Template{
		tmpl:       tmpl,
		flashStore: flashStore,
	}
}

// getTemplateFuncs returns the template functions map
func (t *Template) getTemplateFuncs(r *http.Request) template.FuncMap {
	ctx := r.Context()
	successes := t.flashStore.GetSuccesses(ctx)
	errors := t.flashStore.GetErrors(ctx)
	return template.FuncMap{
		"csrfField": func() template.HTML {
			return csrf.GenerateCsrfField(r)
		},
		"currentUser": func() *model.User {
			user, _ := service.GetUser(ctx)
			return user
		},
		"flashMessages": func() (map[string][]string, error) {
			return map[string][]string{
				"successes": successes,
				"errors":    errors,
			}, nil
		},
	}
}

// GetTmpl returns the underlying template.Template
func (t *Template) Raw() *template.Template {
	return t.tmpl
}

// Render executes a template and writes the output to w
func (t *Template) Render(w http.ResponseWriter, r *http.Request, data any) error {
	// Create a new template with updated funcMap
	newTmpl, err := t.tmpl.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone template: %w", err)
	}

	newTmpl.Funcs(t.getTemplateFuncs(r))

	buf := &bytes.Buffer{}
	// Execute the template with the full template name including layout
	if err := newTmpl.Execute(buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = io.Copy(w, buf)
	return err
}
