package main

import (
	"log"
	"net/http"

	"github.com/shuvo-paul/sitemonitor/cmd/sitemonitor/bootstrap"
	"github.com/shuvo-paul/sitemonitor/internal/app/routes"
)

func main() {
	app := bootstrap.NewApp()
	defer app.Close()
	handler := routes.SetupRoutes(
		app.UserController,
		*app.SessionService,
		*app.UserService,
		app.SiteController,
	)

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
