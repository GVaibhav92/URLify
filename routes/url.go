package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"URLify/models"
	"URLify/services"
	"URLify/utils"
)

// encapsulates dependencies for URL‑related HTTP handlers.
type URLHandler struct {
	urlStore        *models.URLStore
	redis           *redis.Client
	redirectService *services.RedirectService
}

// used in the router setup to bind routes to handler methods.
func NewURLHandler(urlStore *models.URLStore, redis *redis.Client, redirectService *services.RedirectService) *URLHandler {
	return &URLHandler{
		urlStore:        urlStore,
		redis:           redis,
		redirectService: redirectService,
	}
}

type createURLRequest struct {
	OriginalURL string `json:"original_url" validate:"required,url"`
	CustomCode  string `json:"custom_code"`
}

// POST/urls
func (h *URLHandler) CreateURL(c *gin.Context) {
	var req createURLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	uid := userID.(uuid.UUID)

	var shortCode string
	var isCustom bool

	if req.CustomCode != "" {
		if !utils.ValidateCustomCode(req.CustomCode) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Custom code must be 3-30 alphanumeric characters",
			})
			return
		}

		exists, err := h.urlStore.ShortCodeExists(req.CustomCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Custom code already taken"})
			return
		}

		shortCode = req.CustomCode
		isCustom = true

	} else {
		for {
			code, err := utils.GenerateShortCode()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate shortcode"})
				return
			}

			exists, err := h.urlStore.ShortCodeExists(code)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}

			if !exists {
				shortCode = code
				break
			}
			// collision: extremely rare, so loop retries once.
		}
		isCustom = false
	}

	//create the URL in the database.
	url, err := h.urlStore.Create(req.OriginalURL, shortCode, uid, isCustom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create URL"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":      "URL created successfully",
		"id":           url.ID,
		"short_code":   url.ShortCode,
		"original_url": url.OriginalURL,
		"short_url":    "http://localhost:8080/r/" + url.ShortCode,
		"is_custom":    url.IsCustom,
		"created_at":   url.CreatedAt,
	})
}

// GET /urls
func (h *URLHandler) ListURLs(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid := userID.(uuid.UUID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	offset := (page - 1) * limit

	urls, err := h.urlStore.GetByUserID(uid, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch URLs"})
		return
	}

	total, err := h.urlStore.CountByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count URLs"})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"data":        urls,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// DELETE /urls/:id
func (h *URLHandler) DeleteURL(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid := userID.(uuid.UUID)

	// read id parameter from URL path and parse as UUID.
	idParam := c.Param("id")
	urlID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL ID"})
		return
	}

	// fetch the URL record to verify existence and ownership.
	existing, err := h.urlStore.GetByID(urlID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// ownership check
	if existing.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not own this URL"})
		return
	}

	if err := h.urlStore.Delete(urlID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete URL"})
		return
	}

	//Invalidate Redis cache
	_ = h.redirectService.InvalidateCache(c.Request.Context(), existing.ShortCode)

	c.JSON(http.StatusOK, gin.H{"message": "URL deleted successfully"})
}
