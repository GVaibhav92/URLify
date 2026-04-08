package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"URLify/config"
	"URLify/middleware"
	"URLify/models"
)

func Setup(r *gin.Engine, db *sqlx.DB, rdb *redis.Client, cfg *config.Config) {
	userStore := models.NewUserStore(db)
	urlStore := models.NewURLStore(db)

	authHandler := NewAuthHandler(userStore, cfg)
	urlHandler := NewURLHandler(urlStore, rdb)

	// Public routes
	auth := r.Group("/auth")
	{
		auth.POST("/signup", authHandler.Signup)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		protected.POST("/urls", urlHandler.CreateURL)
		protected.GET("/urls", urlHandler.ListURLs)
		protected.DELETE("/urls/:id", urlHandler.DeleteURL)
	}
}
