package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"UTM_Tracker/internal/database"
	"UTM_Tracker/internal/models"
	"UTM_Tracker/internal/tracking"

	"github.com/gin-gonic/gin"
)

type RedirectHandler struct {
	db           *database.Database
	clickTracker *tracking.ClickTracker
}

func NewRedirectHandler(db *database.Database, clickTracker *tracking.ClickTracker) *RedirectHandler {
	return &RedirectHandler{
		db:           db,
		clickTracker: clickTracker,
	}
}

func (h *RedirectHandler) HandleRedirect(c *gin.Context) {
	slug := c.Param("slug")

	// Retrieve URL mapping
	urlMapping, err := h.db.GetURLBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// Log the click
	err = h.clickTracker.LogClick(
		slug,
		c.ClientIP(),
		c.GetHeader("Referer"),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		log.Printf("Failed to log click: %v", err)
	}

	// Construct redirect URL with UTM parameters
	redirectURL := h.buildRedirectURL(urlMapping)

	// Redirect to the destination URL
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *RedirectHandler) buildRedirectURL(mapping *models.URLMapping) string {
	baseURL := mapping.DestinationURL
	utmParams := fmt.Sprintf("utm_source=%s&utm_medium=%s&utm_campaign=%s",
		mapping.UTMSource,
		mapping.UTMMedium,
		mapping.UTMCampaign,
	)

	// Append UTM parameters, handling existing query parameters
	if strings.Contains(baseURL, "?") {
		return baseURL + "&" + utmParams
	}
	return baseURL + "?" + utmParams
}
