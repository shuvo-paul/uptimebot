package csrf

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"net/http"
)

const (
	tokenLength   = 32
	cookieName    = "csrf_token"
	headerName    = "X-CSRF-Token"
	formFieldName = "csrf_token"
)

var errTokenMismatch = errors.New("CSRF token mismatch")

func generateToken() (string, error) {
	token := make([]byte, tokenLength)
	_, err := rand.Read(token)

	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(token), nil
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			if _, err := r.Cookie(cookieName); err != nil {
				setTokenCookie(w)
			}
			next.ServeHTTP(w, r)
			return
		}

		if err := validatedToken(r); err != nil {
			http.Error(w, "Forbidden: CSRF validation failed", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func setTokenCookie(w http.ResponseWriter) {
	token, err := generateToken()
	if err != nil {
		http.Error(w, "Failed to generate CSRF Token", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})
}

func validatedToken(r *http.Request) error {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return err
	}

	token := r.Header.Get(headerName)
	if token == "" {
		token = r.FormValue(formFieldName)
	}

	if token == "" || token != cookie.Value {
		return errTokenMismatch
	}

	return nil
}

func GetToken(r *http.Request) string {
	cookie, err := r.Cookie(cookieName)

	if err != nil {
		return ""
	}

	return cookie.Value
}

func GenerateCsrfField(r *http.Request) template.HTML {
	token := GetToken(r)

	if token == "" {
		return ""
	}

	field := fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, formFieldName, token)

	return template.HTML(field)
}
