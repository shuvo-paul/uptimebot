package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shuvo-paul/uptimebot/internal/bootstrap"
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
	log.Printf("Server starting on :%d", app.Config.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", app.Config.Port), handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
