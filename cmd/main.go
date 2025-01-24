package main

import (
	"log"
	"net/http"

	"github.com/shuvo-paul/uptimebot/cmd/bootstrap"
	"github.com/shuvo-paul/uptimebot/internal/routes"
)

func main() {
	app := bootstrap.NewApp()
	defer app.Close()
	handler := routes.SetupRoutes(
		app.UserHandler,
		*app.SessionService,
		*app.AuthService,
		app.TargetHandler,
		app.NotifierHandler,
	)

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
