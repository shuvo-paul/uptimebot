package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	authService "github.com/shuvo-paul/uptimebot/internal/auth/service"
	targetService "github.com/shuvo-paul/uptimebot/internal/monitor/service"
	"github.com/shuvo-paul/uptimebot/internal/renderer"
	"github.com/shuvo-paul/uptimebot/pkg/flash"
)

type TargetHandler struct {
	targetService targetService.TargetServiceInterface
	flash         flash.FlashStoreInterface
	Template      struct {
		List   *renderer.Template
		Create *renderer.Template
		Edit   *renderer.Template
	}
}

func NewTargetHandler(targetService targetService.TargetServiceInterface, flash flash.FlashStoreInterface) *TargetHandler {
	c := &TargetHandler{
		targetService: targetService,
		flash:         flash,
	}

	return c
}

func (c *TargetHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := authService.GetUser(r.Context())
	if !ok {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	targets, err := c.targetService.GetAllByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch targets", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"title":   "all targets",
		"targets": targets,
	}

	c.Template.List.Render(w, r, data)
}

func (c *TargetHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]any{
			"title": "add a target",
		}
		c.Template.Create.Render(w, r, data)
		return
	}

	url := r.FormValue("url")
	intervalStr := r.FormValue("interval")

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		errors := []string{
			"Invalid interval value",
		}
		c.flash.SetErrors(r.Context(), errors)
		http.Redirect(w, r, "/targets/create", http.StatusSeeOther)
		return
	}

	user, ok := authService.GetUser(r.Context())
	if !ok {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	_, err = c.targetService.Create(user.ID, url, time.Duration(interval)*time.Second)
	if err != nil {
		errors := []string{
			"Failed to create target: " + err.Error(),
		}
		c.flash.SetErrors(r.Context(), errors)
		http.Redirect(w, r, "/targets/create", http.StatusSeeOther)
		return
	}

	successes := []string{
		"Target created successfully",
	}
	c.flash.SetSuccesses(r.Context(), successes)
	http.Redirect(w, r, "/targets", http.StatusSeeOther)
}

func (c *TargetHandler) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid target ID", http.StatusBadRequest)
		slog.Error("Invalid target ID", "error", err)
		return
	}

	if r.Method == http.MethodGet {
		target, err := c.targetService.GetByID(id)
		if err != nil {
			http.Error(w, "Target not found", http.StatusNotFound)
			return
		}

		data := map[string]any{
			"Title":  "Edit Target",
			"target": target,
		}

		c.Template.Edit.Render(w, r, data)
		return
	}

	target, err := c.targetService.GetByID(id)
	if err != nil {
		http.Error(w, "Target not found", http.StatusNotFound)
		return
	}

	target.URL = r.FormValue("url")
	intervalStr := r.FormValue("interval")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		errors := []string{
			"Invalid interval value",
		}
		c.flash.SetErrors(r.Context(), errors)
		http.Redirect(w, r, "/targets/"+strconv.Itoa(id)+"/edit", http.StatusSeeOther)
		return
	}
	target.Interval = time.Duration(interval) * time.Second

	_, err = c.targetService.Update(target)
	if err != nil {
		errors := []string{
			"Failed to update target: " + err.Error(),
		}
		c.flash.SetErrors(r.Context(), errors)
		http.Redirect(w, r, "/targets/"+strconv.Itoa(id)+"/edit", http.StatusSeeOther)
		return
	}

	c.flash.SetSuccesses(r.Context(), []string{"Target updated successfully"})

	http.Redirect(w, r, "/targets", http.StatusSeeOther)
}

func (c *TargetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, "Invalid target ID", http.StatusBadRequest)
		return
	}

	err = c.targetService.Delete(id)
	if err != nil {
		c.flash.SetErrors(r.Context(), []string{"Failed to delete target: " + err.Error()})
	} else {
		c.flash.SetSuccesses(r.Context(), []string{"Target deleted successfully"})
	}

	http.Redirect(w, r, "/targets", http.StatusSeeOther)
}
