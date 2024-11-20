package controllers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/shuvo-paul/sitemonitor/flash"
	"github.com/shuvo-paul/sitemonitor/models"
	"github.com/shuvo-paul/sitemonitor/services"
)

type UserController struct {
	Template struct {
		Register *template.Template
		Login    *template.Template
		Execute  func(w http.ResponseWriter, r *http.Request, tmpl *template.Template, data any)
	}
	sessionService services.SessionServiceInterface
	userService    services.UserServiceInterface
	flashStore     flash.FlashStoreInterface
}

func NewUserController(
	userService services.UserServiceInterface,
	sessionService services.SessionServiceInterface,
	flashStore flash.FlashStoreInterface,
) *UserController {
	return &UserController{
		userService:    userService,
		sessionService: sessionService,
		flashStore:     flashStore,
	}
}

func (c *UserController) ShowRegisterForm(w http.ResponseWriter, r *http.Request) {
	if c.redirectIfAuthenticated(w, r) {
		return
	}

	flashId := flash.GetFlashIDFromContext(r.Context())

	data := map[string]any{
		"Title": "Registration",
		"Erros": c.flashStore.GetFlash(flashId, "errors"),
	}

	c.Template.Execute(w, r, c.Template.Register, data)
}

func (c *UserController) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	user := &models.User{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	_, err := c.userService.CreateUser(user)
	if err != nil {
		flashId := flash.GetFlashIDFromContext(r.Context())

		errors := []string{err.Error()}
		fmt.Println(errors[0])
		c.flashStore.SetFlash(flashId, "errors", errors)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (c *UserController) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	if c.redirectIfAuthenticated(w, r) {
		return
	}

	data := map[string]string{
		"Title": "Login",
	}
	c.Template.Execute(w, r, c.Template.Login, data)
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

func (c *UserController) redirectIfAuthenticated(w http.ResponseWriter, r *http.Request) bool {
	if cookie, err := r.Cookie("session_token"); err == nil {
		user, err := c.sessionService.ValidateSession(cookie.Value)

		if err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			})
			return false
		}
		if user != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return true
		}
	}
	return false
}
