package routes

import (
	"fmt"
	"net/http"

	"github.com/shuvo-paul/sitemonitor/controllers"
	"github.com/shuvo-paul/sitemonitor/middleware"
	"github.com/shuvo-paul/sitemonitor/services"
	"github.com/shuvo-paul/sitemonitor/static"
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

	// Protected routes
	mux.Handle("GET /", middleware.RequireAuth(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("index")
			fmt.Fprint(w, "Hello World")
		}),
		sessionService,
		userService,
	))

	return mux
}
