package routes

import (
	"fmt"
	"net/http"

	"github.com/shuvo-paul/sitemonitor/controllers"
	"github.com/shuvo-paul/sitemonitor/static"
)

func SetupRoutes(userController *controllers.UserController) http.Handler {
	// Setup routes
	mux := http.NewServeMux()

	// OR strip the /static/ prefix (recommended)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.StaticFS))))

	mux.HandleFunc("GET /register", userController.ShowRegisterForm)

	mux.HandleFunc("POST /register", userController.Register)

	mux.HandleFunc("GET /login", userController.ShowLoginForm)

	mux.HandleFunc("POST /login", userController.Login)

	mux.HandleFunc("GET /*", func(w http.ResponseWriter, r *http.Request) {
		// template.Parse("index.html").Execute(w, nil)
		fmt.Println("index")
		fmt.Fprint(w, "Hello World")
	})

	return mux
}
