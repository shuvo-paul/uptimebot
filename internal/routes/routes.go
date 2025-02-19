package routes

import (
	"net/http"

	authHandler "github.com/shuvo-paul/uptimebot/internal/auth/handler"
	authService "github.com/shuvo-paul/uptimebot/internal/auth/service"
	"github.com/shuvo-paul/uptimebot/internal/middleware"
	uptimeHandler "github.com/shuvo-paul/uptimebot/internal/monitor/handler"
	eventHandler "github.com/shuvo-paul/uptimebot/internal/notification/handler"
	"github.com/shuvo-paul/uptimebot/pkg/csrf"
	"github.com/shuvo-paul/uptimebot/pkg/flash"
	"github.com/shuvo-paul/uptimebot/web/static"
)

func SetupRoutes(
	userHandler *authHandler.UserHandler,
	sessionService authService.SessionService,
	authService authService.AuthService,
	targetHandler *uptimeHandler.TargetHandler,
	notifierHandler *eventHandler.NotifierHandler,
) http.Handler {
	// Setup routes
	mux := http.NewServeMux()

	// Public routes
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.StaticFS))))
	mux.HandleFunc("GET /register", userHandler.ShowRegisterForm)
	mux.HandleFunc("POST /register", userHandler.Register)
	mux.HandleFunc("GET /login", userHandler.ShowLoginForm)
	mux.HandleFunc("POST /login", userHandler.Login)
	mux.HandleFunc("POST /logout", userHandler.Logout)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		http.Redirect(w, r, "/targets", http.StatusFound)
	})

	mux.HandleFunc("GET /verification", userHandler.VerifyEmail)
	mux.HandleFunc("POST /verification", userHandler.SendVerificationEmail)

	// Protected routes
	protected := http.NewServeMux()
	// Add target monitoring routes
	protected.HandleFunc("GET /", targetHandler.List)
	protected.HandleFunc("GET /create", targetHandler.Create)
	protected.HandleFunc("POST /create", targetHandler.Create)
	protected.HandleFunc("GET /{id}/edit", targetHandler.Edit)
	protected.HandleFunc("POST /{id}/edit", targetHandler.Edit)
	protected.HandleFunc("POST /{id}/delete", targetHandler.Delete)

	protected.HandleFunc("GET /auth/slack/{targetId}", notifierHandler.AuthSlack)
	protected.HandleFunc("GET /auth/slack/callback", notifierHandler.AuthSlackCallback)

	mux.Handle("/targets/", middleware.RequireAuth(
		http.StripPrefix("/targets", protected),
		sessionService,
		authService,
	))

	mws := middleware.CreateStack(
		flash.Middleware,
		csrf.Middleware,
		middleware.ErrorHandler,
		middleware.Logger,
	)
	return mws(mux)
}
