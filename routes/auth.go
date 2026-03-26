package routes

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"GoSnip/config"
	"GoSnip/models"
	"GoSnip/utils"
)

type AuthHandler struct {
	userStore *models.UserStore
	cfg       *config.Config
}

func NewAuthHandler(userStore *models.UserStore, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userStore: userStore,
		cfg:       cfg,
	}
}

// ── Request body structs ──────────────────────────────────────────────────────

type signupRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func (h *AuthHandler) Signup(c *gin.Context) {
	var req signupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	existing, err := h.userStore.GetByEmail(req.Email)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user, err := h.userStore.Create(req.Email, hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Role, h.cfg.JWTSecret, h.cfg.JWTExpiryHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"token":   token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userStore.GetByEmail(req.Email)
	if err != nil {
		// Intentionally vague — don't reveal whether email exists
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Role, h.cfg.JWTSecret, h.cfg.JWTExpiryHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}
