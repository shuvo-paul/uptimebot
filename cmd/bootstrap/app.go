package bootstrap

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	authHandler "github.com/shuvo-paul/uptimebot/internal/auth/handler"
	authRepository "github.com/shuvo-paul/uptimebot/internal/auth/repository"
	authService "github.com/shuvo-paul/uptimebot/internal/auth/service"
	"github.com/shuvo-paul/uptimebot/internal/database"
	"github.com/shuvo-paul/uptimebot/internal/database/migrations"
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

var db *sql.DB

type App struct {
	AuthService     *authService.AuthService
	SessionService  *authService.SessionService
	UserHandler     *authHandler.UserHandler
	TargetHandler   *uptimeHandler.TargetHandler
	NotifierHandler *notificationHandler.NotifierHandler
}

func NewApp() *App {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	db, err := database.InitDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	migrations.SetupMigration(db)

	templateRenderer := renderer.New(templates.TemplateFS)

	flashStore := flash.NewFlashStore()

	userRepository := authRepository.NewUserRepository(db)
	sessionRepository := authRepository.NewSessionRepository(db)

	authService2 := authService.NewAuthService(userRepository)

	sessionService := authService.NewSessionService(sessionRepository)
	authHandler := authHandler.NewUserHandler(authService2, sessionService, flashStore)
	authHandler.Template.Register = templateRenderer.GetTemplate("pages:register")
	authHandler.Template.Login = templateRenderer.GetTemplate("pages:login")

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
		AuthService:     authService2,
		SessionService:  sessionService,
		UserHandler:     authHandler,
		TargetHandler:   targetHandler,
		NotifierHandler: notifierHandler,
	}
}

func (a *App) Close() {
	db.Close()
}
