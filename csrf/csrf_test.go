package csrf

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token1, err1 := generateToken()
	token2, err2 := generateToken()

	if err1 != nil || err2 != nil {
		t.Errorf("generateToken() failed: %v, %v", err1, err2)
	}

	if token1 == "" || token2 == "" {
		t.Error("generateToken() returned empty token")
	}

	if token1 == token2 {
		t.Error("generateToken() returned identical tokens")
	}

	if len(token1) != 43 { // base64 encoded 32 bytes
		t.Errorf("unexpected token length: got %d, want 43", len(token1))
	}
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		setCookie      bool
		setHeader      bool
		setFormValue   bool
		expectedStatus int
	}{
		{
			name:           "GET request without cookie",
			method:         http.MethodGet,
			setCookie:      false,
			setHeader:      false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request without token",
			method:         http.MethodPost,
			setCookie:      true,
			setHeader:      false,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "POST request with valid token",
			method:         http.MethodPost,
			setCookie:      true,
			setHeader:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request with valid form token",
			method:         http.MethodPost,
			setCookie:      true,
			setHeader:      false,
			setFormValue:   true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/", nil)
			rec := httptest.NewRecorder()

			if tt.setCookie {
				token, _ := generateToken()
				req.AddCookie(&http.Cookie{
					Name:  cookieName,
					Value: token,
				})
				if tt.setHeader {
					req.Header.Set(headerName, token)
				}
				if tt.setFormValue {
					form := url.Values{}
					form.Add(formFieldName, token)
					req.PostForm = form
				}
			}

			Middleware(handler).ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestGetToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	token := "test-token"
	req.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: token,
	})

	if got := GetToken(req); got != token {
		t.Errorf("GetToken() = %v, want %v", got, token)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	if got := GetToken(req); got != "" {
		t.Errorf("GetToken() with no cookie = %v, want empty string", got)
	}
}

func TestGenerateCsrfField(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	token := "test-token"
	req.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: token,
	})

	expected := `<input type="hidden" name="csrf_token" value="test-token">`
	if got := GenerateCsrfField(req); string(got) != expected {
		t.Errorf("GenerateCsrfField() = %v, want %v", got, expected)
	}
}
