package middleware

import (
	"net/http"
	"strings"
)

// RemoveTrailingSlash is a middleware that removes trailing slashes from URLs
// except for the root path "/". This helps normalize URLs and reduce route duplication.
func RemoveTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		next.ServeHTTP(w, r)
	})
}
