package renderer

import (
	"embed"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/auth/service"
	"github.com/shuvo-paul/uptimebot/internal/renderer/testdata"
	"github.com/stretchr/testify/assert"
)

var testFS = testdata.FS

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		fs      embed.FS
		wantErr bool
	}{
		{
			name:    "valid filesystem with templates",
			fs:      testFS,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New(tt.fs)
			assert.NotNil(t, engine)
			assert.NotEmpty(t, engine.templates)
			assert.NotEmpty(t, engine.layouts)
		})
	}
}

func TestEngine_parseAllTemplates(t *testing.T) {
	engine := &Engine{
		fs:        testFS,
		templates: make(map[string]*Template),
		layouts:   make([]string, 0),
	}

	// Add template functions for testing
	tmpl := template.New("").Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return template.HTML(`<input type="hidden" name="csrf_token" value="test-token">`)
		},
		"currentUser": func() *model.User {
			return nil
		},
	})

	// Set the template functions
	engine.templates["test"] = &Template{tmpl: tmpl}

	err := engine.parseAllTemplates()
	assert.NoError(t, err)
	assert.NotEmpty(t, engine.layouts)
	assert.NotEmpty(t, engine.templates)
}

func TestEngine_GetTemplate(t *testing.T) {
	engine := New(testFS)

	tests := []struct {
		name      string
		key       string
		wantPanic bool
	}{
		{
			name:      "existing template",
			key:       "pages:index",
			wantPanic: false,
		},
		{
			name:      "non-existent template",
			key:       "pages:nonexistent",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic {
					assert.NotNil(t, r)
					assert.Contains(t, r.(string), "template not found")
				} else {
					assert.Nil(t, r)
				}
			}()

			tmpl := engine.GetTemplate(tt.key)
			if !tt.wantPanic {
				assert.NotNil(t, tmpl)
			}
		})
	}
}

func TestTemplate_Render(t *testing.T) {
	engine := New(testFS)

	tests := []struct {
		name     string
		key      string
		data     interface{}
		setUser  bool
		wantCode int
		wantBody string
	}{
		{
			name:     "successful render",
			key:      "pages:index",
			data:     map[string]string{"title": "This is a title", "Name": "Test User"},
			setUser:  false,
			wantCode: http.StatusOK,
			wantBody: "Welcome Test User",
		},
		{
			name:     "successful render with user context",
			key:      "pages:index",
			data:     map[string]string{"title": "This is a title", "Name": "Test User"},
			setUser:  true,
			wantCode: http.StatusOK,
			wantBody: "Welcome Test User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			// Set user context if required
			if tt.setUser {
				ctx := service.WithUser(r.Context(), &model.User{
					ID:    1,
					Email: "test@example.com",
				})
				r = r.WithContext(ctx)
			}

			tmpl := engine.GetTemplate(tt.key)
			assert.NotNil(t, tmpl)

			err := tmpl.Render(w, r, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
