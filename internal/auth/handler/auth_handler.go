package handler

import (
	"net/http"
	"strconv"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/auth/service"
	"github.com/shuvo-paul/uptimebot/internal/renderer"
	"github.com/shuvo-paul/uptimebot/pkg/flash"
)

type AuthHandler struct {
	Template struct {
		Register             *renderer.Template
		Login                *renderer.Template
		ResetPassword        *renderer.Template
		RequestPasswordReset *renderer.Template
		Profile              *renderer.Template
	}
	sessionService service.SessionServiceInterface
	authService    service.AuthServiceInterface
	flashStore     flash.FlashStoreInterface
}

func NewAuthHandler(
	authService service.AuthServiceInterface,
	sessionService service.SessionServiceInterface,
	flashStore flash.FlashStoreInterface,
) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		sessionService: sessionService,
		flashStore:     flashStore,
	}
}

func (c *AuthHandler) ShowRegisterForm(w http.ResponseWriter, r *http.Request) {
	if c.redirectIfAuthenticated(w, r) {
		return
	}

	data := map[string]any{
		"Title": "Registration",
	}

	c.Template.Register.Render(w, r, data)
}

func (c *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	user := &model.User{
		
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	_, err := c.authService.CreateUser(user)
	if err != nil {
		errors := []string{err.Error()}
		c.flashStore.SetErrors(ctx, errors)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	c.flashStore.SetSuccesses(ctx, []string{"Registration successful! Please check your email to verify your account."})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (c *AuthHandler) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	if c.redirectIfAuthenticated(w, r) {
		return
	}

	data := map[string]any{
		"Title": "Login",
	}
	c.Template.Login.Render(w, r, data)
}

func (c *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := c.authService.Authenticate(email, password)
	if err != nil {
		c.flashStore.SetErrors(ctx, []string{err.Error()})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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

	http.Redirect(w, r, "/targets", http.StatusSeeOther)
}

func (c *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := r.URL.Query().Get("token")

	if token == "" {
		http.Error(w, "Missing verification token", http.StatusBadRequest)
		return
	}

	err := c.authService.VerifyEmail(token)
	if err != nil {
		c.flashStore.SetErrors(ctx, []string{"Invalid or expired verification token"})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	c.flashStore.SetSuccesses(ctx, []string{"Email verified successfully!"})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (c *AuthHandler) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.FormValue("user_id")
	if userID == "" {
		c.flashStore.SetErrors(ctx, []string{"Missing user ID"})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.flashStore.SetErrors(ctx, []string{"Invalid user ID"})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := c.authService.GetUserByID(userIDInt)
	if err != nil {
		c.flashStore.SetErrors(ctx, []string{"Failed to get user"})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = c.authService.SendToken(user.ID, user.Email, model.TokenTypeEmailVerification)
	if err != nil {
		c.flashStore.SetErrors(ctx, []string{"Failed to send verification email"})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	c.flashStore.SetSuccesses(ctx, []string{"Verification email resent successfully!"})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (c *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Invalidate the session in the backend
		if err := c.sessionService.DeleteSession(cookie.Value); err != nil {
			http.Error(w, "Failed to logout", http.StatusInternalServerError)
			return
		}
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (c *AuthHandler) ShowRequestResetForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Send Reset Password Email",
	}
	c.Template.RequestPasswordReset.Render(w, r, data)
}

func (c *AuthHandler) SendResetLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		c.flashStore.SetErrors(ctx, []string{"Email is required"})
		http.Redirect(w, r, "/request-reset-password", http.StatusSeeOther)
		return
	}

	user, err := c.authService.GetUserByEmail(email)
	if err != nil {
		// Don't reveal if email exists or not for security
		c.flashStore.SetSuccesses(ctx, []string{"If your email exists in our system, you will receive a password reset link shortly."})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = c.authService.SendToken(user.ID, email, model.TokenTypePasswordReset)
	if err != nil {
		c.flashStore.SetErrors(ctx, []string{"Failed to send reset password email"})
		http.Redirect(w, r, "/request-reset-password", http.StatusSeeOther)
		return
	}

	c.flashStore.SetSuccesses(ctx, []string{"Password reset link has been sent to your email"})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (c *AuthHandler) ShowResetPasswordForm(w http.ResponseWriter, r *http.Request) {
	if c.redirectIfAuthenticated(w, r) {
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	_, err := c.authService.ValidateToken(token, model.TokenTypePasswordReset)
	if err != nil {
		c.flashStore.SetErrors(r.Context(), []string{"Invalid or expired token"})
		http.Redirect(w, r, "/request-reset-password", http.StatusSeeOther)
		return
	}

	data := map[string]any{
		"Title": "Reset Password",
		"Token": token,
	}
	c.Template.ResetPassword.Render(w, r, data)
}

func (c *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	token := r.FormValue("token")
	password := r.FormValue("password")
	conformPassword := r.FormValue("confirm_password")
	if password != conformPassword {
		c.flashStore.SetErrors(ctx, []string{"Passwords do not match"})
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusSeeOther)
		return
	}
	if err := c.authService.ResetPassword(token, password); err != nil {
		c.flashStore.SetErrors(ctx, []string{"Failed to reset password"})
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusSeeOther)
		return
	}

	c.flashStore.SetSuccesses(ctx, []string{"Password reset successfully!"})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (c *AuthHandler) ShowProfileForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Profile",
	}
	c.Template.Profile.Render(w, r, data)
}

func (c *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	user, ok := service.GetUser(ctx)
	if !ok {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	if newPassword != confirmPassword {
		c.flashStore.SetErrors(ctx, []string{"New passwords do not match"})
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Verify current password
	_, err := c.authService.Authenticate(user.Email, currentPassword)
	if err != nil {
		c.flashStore.SetErrors(ctx, []string{"Current password is incorrect"})
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Update password
	if err := c.authService.UpdatePassword(user.ID, newPassword); err != nil {
		c.flashStore.SetErrors(ctx, []string{err.Error()})
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	c.flashStore.SetSuccesses(ctx, []string{"Password updated successfully"})
	http.Redirect(w, r, "/targets", http.StatusSeeOther)
}

func (c *AuthHandler) redirectIfAuthenticated(w http.ResponseWriter, r *http.Request) bool {
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
			http.Redirect(w, r, "/targets", http.StatusSeeOther)
			return true
		}
	}
	return false
}
