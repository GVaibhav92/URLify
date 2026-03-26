package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"GoSnip/config"
	"GoSnip/models"
)

func Setup(r *gin.Engine, db *sqlx.DB, cfg *config.Config) {
	userStore := models.NewUserStore(db)
	authHandler := NewAuthHandler(userStore, cfg)

	// Auth routes — public
	auth := r.Group("/auth")
	{
		auth.POST("/signup", authHandler.Signup)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes come in Phase 3
}
