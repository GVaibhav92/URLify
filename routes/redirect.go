package routes

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"URLify/metrics"
	"URLify/services"
)

type RedirectHandler struct {
	redirectService *services.RedirectService
}

func NewRedirectHandler(redirectService *services.RedirectService) *RedirectHandler {
	return &RedirectHandler{redirectService: redirectService}
}

func (h *RedirectHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("shortcode")

	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code required"})
		return
	}

	originalURL, fromCache, err := h.redirectService.Resolve(c.Request.Context(), shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve URL"})
		return
	}

	if fromCache {
		c.Header("X-Cache", "HIT")
		metrics.RedirectCacheHits.Inc()
	} else {
		c.Header("X-Cache", "MISS")
		metrics.RedirectCacheMisses.Inc()
	}
	metrics.RedirectsTotal.Inc()

	c.Redirect(http.StatusMovedPermanently, originalURL)
}
