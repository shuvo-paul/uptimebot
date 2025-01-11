package controllers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/shuvo-paul/sitemonitor/internal/app/renderer"
	"github.com/shuvo-paul/sitemonitor/internal/app/services"
	"github.com/shuvo-paul/sitemonitor/pkg/flash"
)

type SiteController struct {
	siteService services.SiteServiceInterface
	flash       flash.FlashStoreInterface
	Template    struct {
		List   renderer.PageTemplate
		Create renderer.PageTemplate
		Edit   renderer.PageTemplate
	}
}

func NewSiteController(siteService services.SiteServiceInterface, flash flash.FlashStoreInterface) *SiteController {
	c := &SiteController{
		siteService: siteService,
		flash:       flash,
	}

	return c
}

func (c *SiteController) List(w http.ResponseWriter, r *http.Request) {
	user, ok := services.GetUser(r.Context())
	if !ok {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	sites, err := c.siteService.GetAllByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch sites", http.StatusInternalServerError)
		return
	}

	flashId := flash.GetFlashIDFromContext(r.Context())

	data := map[string]any{
		"title":   "all sites",
		"sites":   sites,
		"success": c.flash.GetFlash(flashId, "success"),
	}

	c.Template.List.Render(w, r, data)
}

func (c *SiteController) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]any{
			"title": "add a sites",
		}
		c.Template.Create.Render(w, r, data)
		return
	}

	url := r.FormValue("url")
	intervalStr := r.FormValue("interval")

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		flashID := flash.GetFlashIDFromContext(r.Context())
		c.flash.SetFlash(flashID, "error", "Invalid interval value")
		http.Redirect(w, r, "/sites/create", http.StatusSeeOther)
		return
	}

	user, ok := services.GetUser(r.Context())
	if !ok {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	_, err = c.siteService.Create(user.ID, url, time.Duration(interval)*time.Second)
	if err != nil {
		flashID := flash.GetFlashIDFromContext(r.Context())
		c.flash.SetFlash(flashID, "error", "Failed to create site: "+err.Error())
		http.Redirect(w, r, "/sites/create", http.StatusSeeOther)
		return
	}

	flashID := flash.GetFlashIDFromContext(r.Context())
	c.flash.SetFlash(flashID, "success", "Site created successfully")
	http.Redirect(w, r, "/sites", http.StatusSeeOther)
}

func (c *SiteController) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid site ID", http.StatusBadRequest)
		slog.Error("Invalid site ID", "error", err)
		return
	}

	if r.Method == http.MethodGet {
		site, err := c.siteService.GetByID(id)
		if err != nil {
			http.Error(w, "Site not found", http.StatusNotFound)
			return
		}

		data := map[string]any{
			"Title": "Edit Site",
			"site":  site,
		}

		c.Template.Edit.Render(w, r, data)
		return
	}

	site, err := c.siteService.GetByID(id)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	site.URL = r.FormValue("url")
	intervalStr := r.FormValue("interval")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		flashID := flash.GetFlashIDFromContext(r.Context())
		c.flash.SetFlash(flashID, "error", "Invalid interval value")
		http.Redirect(w, r, "/sites/"+strconv.Itoa(id)+"/edit", http.StatusSeeOther)
		return
	}
	site.Interval = time.Duration(interval) * time.Second

	_, err = c.siteService.Update(site)
	if err != nil {
		flashID := flash.GetFlashIDFromContext(r.Context())
		c.flash.SetFlash(flashID, "error", "Failed to update site: "+err.Error())
		http.Redirect(w, r, "/sites/"+strconv.Itoa(id)+"/edit", http.StatusSeeOther)
		return
	}

	flashID := flash.GetFlashIDFromContext(r.Context())
	c.flash.SetFlash(flashID, "success", "Site updated successfully")
	http.Redirect(w, r, "/sites", http.StatusSeeOther)
}

func (c *SiteController) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, "Invalid site ID", http.StatusBadRequest)
		return
	}

	err = c.siteService.Delete(id)
	flashID := flash.GetFlashIDFromContext(r.Context())
	if err != nil {
		c.flash.SetFlash(flashID, "error", "Failed to delete site: "+err.Error())
	} else {
		c.flash.SetFlash(flashID, "success", "Site deleted successfully")
	}

	http.Redirect(w, r, "/sites", http.StatusSeeOther)
}
