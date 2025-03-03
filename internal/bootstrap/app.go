package bootstrap

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"

	"github.com/joho/godotenv"
	authHandler "github.com/shuvo-paul/uptimebot/internal/auth/handler"
	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	authRepository "github.com/shuvo-paul/uptimebot/internal/auth/repository"
	authService "github.com/shuvo-paul/uptimebot/internal/auth/service"
	"github.com/shuvo-paul/uptimebot/internal/config"
	"github.com/shuvo-paul/uptimebot/internal/database"
	"github.com/shuvo-paul/uptimebot/internal/database/migrations"
	"github.com/shuvo-paul/uptimebot/internal/email"
	uptimeHandler "github.com/shuvo-paul/uptimebot/internal/monitor/handler"
	uptimeRepository "github.com/shuvo-paul/uptimebot/internal/monitor/repository"
	uptimeService "github.com/shuvo-paul/uptimebot/internal/monitor/service"
	notificationHandler "github.com/shuvo-paul/uptimebot/internal/notification/handler"
	notificationRepository "github.com/shuvo-paul/uptimebot/internal/notification/repository"
	notificationService "github.com/shuvo-paul/uptimebot/internal/notification/service"
	"github.com/shuvo-paul/uptimebot/internal/renderer"
	"github.com/shuvo-paul/uptimebot/internal/templates"
	"github.com/shuvo-paul/uptimebot/pkg/flash"
)

type App struct {
	Config          *config.Config
	AuthService     *authService.AuthService
	SessionService  *authService.SessionService
	UserHandler     *authHandler.AuthHandler
	TargetHandler   *uptimeHandler.TargetHandler
	NotifierHandler *notificationHandler.NotifierHandler
	db              *sql.DB
	tempDBDir       *database.TempDBDir
}

func NewApp() *App {
	cfg, err := config.Load()
	if err != nil {
		// Try loading from .env file if config loading fails
		if envErr := godotenv.Load(); envErr != nil {
			log.Printf("Warning: .env file not found or error loading it: %v", envErr)
		}
		// Attempt to load config again after loading .env
		cfg, err = config.Load()
	}
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	tempDBDir, err := database.NewTempDir()

	if err != nil {
		log.Fatalf("Failed to create temporary directory: %v", err)
	}

	db, err := database.InitDatabase(cfg.Database, tempDBDir)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	migrations.SetupMigration(db)

	flashStore := flash.NewFlashStore()

	templateRenderer := renderer.New(templates.TemplateFS, flashStore)

	userRepository := authRepository.NewUserRepository(db)
	sessionRepository := authRepository.NewSessionRepository(db)
	tokenRepository := authRepository.NewTokenRepository(db)

	emailService, err := email.NewEmailService(&cfg.Email)
	if err != nil {
		log.Fatalf("Failed to initialize email service: %v", err)
	}

	emailVerificationTemplate := templateRenderer.GetTemplate("emails:verify-email")
	passwordResetTemplate := templateRenderer.GetTemplate("emails:request-reset-password")
	// Create a map of email templates
	emailTemplates := map[model.TokenType]*template.Template{
		model.TokenTypeEmailVerification: emailVerificationTemplate.Raw(),
		model.TokenTypePasswordReset:     passwordResetTemplate.Raw(),
	}

	// Initialize account token service
	tokenService := authService.NewTokenService(
		tokenRepository,
		emailService,
		cfg.BaseURL,
		emailTemplates,
	)

	// Initialize auth service with token service
	authService2 := authService.NewAuthService(userRepository, tokenService)

	sessionService := authService.NewSessionService(sessionRepository)
	authHandler := authHandler.NewAuthHandler(authService2, sessionService, flashStore)
	authHandler.Template.Register = templateRenderer.GetTemplate("pages:register")
	authHandler.Template.Login = templateRenderer.GetTemplate("pages:login")
	authHandler.Template.RequestPasswordReset = templateRenderer.GetTemplate("pages:request-reset-password")
	authHandler.Template.ResetPassword = templateRenderer.GetTemplate("pages:reset-password")
	authHandler.Template.Profile = templateRenderer.GetTemplate("pages:profile")

	notifierRepository := notificationRepository.NewNotifierRepository(db)
	notifierService := notificationService.NewNotifierService(notifierRepository, nil)
	notifierHandler := notificationHandler.NewNotifierHandler(notifierService)

	targetRepository := uptimeRepository.NewTargetRepository(db)
	targetService := uptimeService.NewTargetService(targetRepository, notifierService)

	// Initialize monitoring for existing targets
	if err := targetService.InitializeMonitoring(); err != nil {
		log.Printf("Failed to initialize target monitoring: %v", err)
		// Don't fatal here, allow the app to continue even if some monitors fail
	}

	// Initialize target controller
	targetHandler := uptimeHandler.NewTargetHandler(targetService, flashStore)
	targetHandler.Template.List = templateRenderer.GetTemplate("pages:targets/list")
	targetHandler.Template.Create = templateRenderer.GetTemplate("pages:targets/create")
	targetHandler.Template.Edit = templateRenderer.GetTemplate("pages:targets/edit")

	fmt.Println("app initialized")

	return &App{
		Config:          cfg,
		AuthService:     authService2,
		SessionService:  sessionService,
		UserHandler:     authHandler,
		TargetHandler:   targetHandler,
		NotifierHandler: notifierHandler,
		db:              db,
		tempDBDir:       tempDBDir,
	}
}

func (a *App) Close() {
	a.db.Close()
	a.tempDBDir.Cleanup()
}
