package bootstrap

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/shuvo-paul/sitemonitor/internal/app/controllers"
	"github.com/shuvo-paul/sitemonitor/internal/app/renderer"
	"github.com/shuvo-paul/sitemonitor/internal/app/repository"
	"github.com/shuvo-paul/sitemonitor/internal/app/services"
	"github.com/shuvo-paul/sitemonitor/internal/database"
	"github.com/shuvo-paul/sitemonitor/internal/database/migrations"
	"github.com/shuvo-paul/sitemonitor/pkg/flash"
	"github.com/shuvo-paul/sitemonitor/web/templates"
)

var db *sql.DB

type App struct {
	UserService        *services.UserService
	SessionService     *services.SessionService
	UserController     *controllers.UserController
	SiteController     *controllers.SiteController
	NotifierController *controllers.NotifierController
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

	userRepository := repository.NewUserRepository(db)
	sessionRepository := repository.NewSessionRepository(db)

	userService := services.NewUserService(userRepository)
	sessionService := services.NewSessionService(sessionRepository)
	userController := controllers.NewUserController(userService, sessionService, flashStore)
	userController.Template.Register = templateRenderer.Parse("register.html")
	userController.Template.Login = templateRenderer.Parse("login.html")

	siteRepository := repository.NewSiteRepository(db)
	siteService := services.NewSiteService(siteRepository)

	// Initialize monitoring for existing sites
	if err := siteService.InitializeMonitoring(); err != nil {
		log.Printf("Failed to initialize site monitoring: %v", err)
		// Don't fatal here, allow the app to continue even if some monitors fail
	}

	// Initialize site controller
	siteController := controllers.NewSiteController(siteService, flashStore)
	siteController.Template.List = templateRenderer.Parse("sites/list.html")
	siteController.Template.Create = templateRenderer.Parse("sites/create.html")
	siteController.Template.Edit = templateRenderer.Parse("sites/edit.html")

	notifierRepository := repository.NewNotifierRepository(db)
	notifierService := services.NewNotifierService(notifierRepository, nil)
	notifierController := controllers.NewNotifierController(notifierService)

	fmt.Println("app initialized")

	return &App{
		UserService:        userService,
		SessionService:     sessionService,
		UserController:     userController,
		SiteController:     siteController,
		NotifierController: notifierController,
	}
}

func (a *App) Close() {
	db.Close()
}
