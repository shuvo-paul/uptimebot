package controllers

import (
	"net/http"
	"text/template"

	"github.com/shuvo-paul/sitemonitor/models"
	"github.com/shuvo-paul/sitemonitor/services"
)

type UserController struct {
	Template struct {
		Register *template.Template
		Login    *template.Template
	}
	sessionService services.SessionServiceInterface
	userService    services.UserServiceInterface
}

func NewUserController(userService services.UserServiceInterface, sessionService services.SessionServiceInterface) *UserController {
	return &UserController{userService: userService, sessionService: sessionService}
}

func (c *UserController) ShowRegisterForm(w http.ResponseWriter, r *http.Request) {
	c.Template.Register.Execute(w, nil)
}

func (c *UserController) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	user := &models.User{
		Username: r.FormValue("username"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	_, err := c.userService.CreateUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (c *UserController) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	c.Template.Login.Execute(w, nil)
}

func (c *UserController) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := c.userService.Authenticate(email, password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	session, token, err := c.sessionService.CreateSession(user.ID)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
