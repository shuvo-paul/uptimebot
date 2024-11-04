package app

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/shuvo-paul/sitemonitor/controllers"
	"github.com/shuvo-paul/sitemonitor/database"
	"github.com/shuvo-paul/sitemonitor/repository"
	"github.com/shuvo-paul/sitemonitor/services"
	"github.com/shuvo-paul/sitemonitor/views"
	"github.com/shuvo-paul/sitemonitor/views/templates"
)

var db *sql.DB

type App struct {
	userService    *services.UserService
	sessionService *services.SessionService
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

	template := views.NewTemplate(templates.TemplateFS)

	userRepository := repository.NewUserRepository(db)
	sessionRepository := repository.NewSessionRepository(db)

	userService := services.NewUserService(userRepository)
	sessionService := services.NewSessionService(sessionRepository)
	userController := controllers.NewUserController(userService, sessionService)
	userController.Template.Register = template.Parse("register.html")
	userController.Template.Login = template.Parse("login.html")
	fmt.Println("app initialized")

	return &App{userService: userService, sessionService: sessionService, UserController: userController}
}

func (a *App) Close() {
	db.Close()
}
