package main

import (
	"log"
	"net/http"

	"github.com/shuvo-paul/sitemonitor/app"
	"github.com/shuvo-paul/sitemonitor/routes"
)

func main() {
	app := app.NewApp()
	defer app.Close()
	handler := routes.SetupRoutes(app.UserController, *app.SessionService, *app.UserService)

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
