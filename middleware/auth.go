package middleware

import (
	"net/http"

	"github.com/shuvo-paul/sitemonitor/services"
)

func RequireAuth(next http.Handler, sessionService services.SessionService, userService services.UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "login", http.StatusSeeOther)
			return
		}

		session, err := sessionService.ValidateSession(cookie.Value)
		if err != nil || session == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := userService.GetUserByID(session.ID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		ctx := services.WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
