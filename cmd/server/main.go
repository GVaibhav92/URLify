package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"URLify/config"
	"URLify/db"
	"URLify/routes"
)

func main() {
	cfg := config.Load()

	pgDB := db.NewPostgres(cfg)
	defer pgDB.Close()

	redisClient := db.NewRedis(cfg)
	defer redisClient.Close()

	r := gin.Default()

	routes.Setup(r, pgDB, redisClient, cfg)

	log.Printf("URLify running on port %s", cfg.AppPort)

	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
