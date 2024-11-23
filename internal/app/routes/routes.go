package routes

import (
	"fmt"
	"net/http"

	"github.com/shuvo-paul/sitemonitor/internal/app/controllers"
	"github.com/shuvo-paul/sitemonitor/internal/app/middleware"
	"github.com/shuvo-paul/sitemonitor/internal/app/services"
	"github.com/shuvo-paul/sitemonitor/pkg/csrf"
	"github.com/shuvo-paul/sitemonitor/pkg/flash"
	"github.com/shuvo-paul/sitemonitor/web/static"
)

func SetupRoutes(
	userController *controllers.UserController,
	sessionService services.SessionService,
	userService services.UserService,
) http.Handler {
	// Setup routes
	mux := http.NewServeMux()

	// Public routes
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.StaticFS))))
	mux.HandleFunc("GET /register", userController.ShowRegisterForm)
	mux.HandleFunc("POST /register", userController.Register)
	mux.HandleFunc("GET /login", userController.ShowLoginForm)
	mux.HandleFunc("POST /login", userController.Login)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		fmt.Fprint(w, "Wellcome")
	})

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Protected Area")
	})

	mux.Handle("/app/", middleware.RequireAuth(
		http.StripPrefix("/app", protected),
		sessionService,
		userService,
	))

	mws := middleware.CreateStack(
		flash.Middleware,
		csrf.Middleware,
		middleware.ErrorHandler,
		middleware.Logger,
	)
	return mws(mux)
}
