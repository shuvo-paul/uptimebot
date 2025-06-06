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
	userHandler *authHandler.AuthHandler,
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
		http.Redirect(w, r, "/app/targets", http.StatusFound)
	})

	mux.HandleFunc("GET /verify-email", userHandler.VerifyEmail)
	mux.HandleFunc("GET /request-reset-password", userHandler.ShowRequestResetForm)
	mux.HandleFunc("POST /send-reset-password-link", userHandler.SendResetLink)
	mux.HandleFunc("GET /reset-password", userHandler.ShowResetPasswordForm)
	mux.HandleFunc("POST /reset-password", userHandler.ResetPassword)

	// Protected routes
	protected := http.NewServeMux()
	// Add target monitoring routes
	protected.HandleFunc("GET /", targetHandler.List)
	// Register specific routes first
	protected.HandleFunc("GET /targets/create", targetHandler.Create)
	protected.HandleFunc("POST /targets/create", targetHandler.Create)

	// Then register wildcard routes
	protected.HandleFunc("GET /targets/edit/{id}", targetHandler.Edit)
	protected.HandleFunc("POST /targets/edit/{id}", targetHandler.Edit)
	protected.HandleFunc("POST /targets/delete/{id}", targetHandler.Delete)
	protected.HandleFunc("POST /targets/toggle-enable/{id}", targetHandler.ToggleEnabled)

	protected.HandleFunc("GET /auth/slack/{targetId}", notifierHandler.AuthSlack)
	protected.HandleFunc("POST /verify-email", userHandler.SendVerificationEmail)
	protected.HandleFunc("POST /profile", userHandler.ShowProfileForm)
	protected.HandleFunc("POST /update-password", userHandler.UpdatePassword)

	// Move Slack callback route to main mux to preserve query parameters
	protected.HandleFunc("GET /auth/slack/callback", notifierHandler.AuthSlackCallback)

	mux.Handle("/app/", middleware.RequireAuth(
		http.StripPrefix("/app", protected),
		sessionService,
		authService,
	))

	mws := middleware.CreateStack(
		flash.Middleware,
		csrf.Middleware,
		middleware.ErrorHandler,
		middleware.Logger,
		middleware.RemoveTrailingSlash,
	)
	return mws(mux)
}
