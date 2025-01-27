package renderer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shuvo-paul/uptimebot/internal/templates"
)

func TestNew(t *testing.T) {
	engine := New(templates.TemplateFS)

	if engine == nil {
		t.Fatal("New() returned nil")
	}

	if engine.fs != templates.TemplateFS {
		t.Error("filesystem not properly set")
	}

	if len(engine.pages) == 0 {
		t.Error("no pages found")
	}

	if engine.funcMap == nil {
		t.Error("funcMap not initialized")
	}
}

func TestParse(t *testing.T) {
	engine := New(templates.TemplateFS)

	tests := []struct {
		name        string
		files       string
		shouldPanic bool
	}{
		{
			name:        "empty files",
			files:       "",
			shouldPanic: true,
		},
		{
			name:        "valid template",
			files:       "test.html",
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.shouldPanic {
					t.Errorf("Parse() panic = %v, shouldPanic = %v", r, tt.shouldPanic)
				}
			}()

			_ = engine.Parse(tt.files)
		})
	}
}

func TestPageTemplate_Render(t *testing.T) {
	engine := New(templates.TemplateFS)
	tmpl := engine.Parse("test.html")

	tests := []struct {
		name     string
		data     any
		wantCode int
		wantBody string
	}{
		{
			name:     "successful render",
			data:     map[string]string{"title": "Test Page"},
			wantCode: http.StatusOK,
			wantBody: "Test Page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			tmpl.Render(w, r, tt.data)

			if w.Code != tt.wantCode {
				t.Errorf("Render() status code = %v, want %v", w.Code, tt.wantCode)
			}

			if !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Render() body = %v, want to contain %v", w.Body.String(), tt.wantBody)
			}
		})
	}
}
