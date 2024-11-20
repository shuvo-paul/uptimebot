package app

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/shuvo-paul/sitemonitor/controllers"
	"github.com/shuvo-paul/sitemonitor/database"
	"github.com/shuvo-paul/sitemonitor/migrations"
	"github.com/shuvo-paul/sitemonitor/repository"
	"github.com/shuvo-paul/sitemonitor/services"
	"github.com/shuvo-paul/sitemonitor/views"
	"github.com/shuvo-paul/sitemonitor/views/templates"
)

var db *sql.DB

type App struct {
	UserService    *services.UserService
	SessionService *services.SessionService
	UserController *controllers.UserController
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

	tpl := views.NewTemplate(templates.TemplateFS)

	userRepository := repository.NewUserRepository(db)
	sessionRepository := repository.NewSessionRepository(db)

	userService := services.NewUserService(userRepository)
	sessionService := services.NewSessionService(sessionRepository)
	userController := controllers.NewUserController(userService, sessionService)
	userController.Template.Register = tpl.Parse("register.html")
	userController.Template.Login = tpl.Parse("login.html")
	userController.Template.Execute = tpl.Execute
	fmt.Println("app initialized")

	return &App{UserService: userService, SessionService: sessionService, UserController: userController}
}

func (a *App) Close() {
	db.Close()
}
