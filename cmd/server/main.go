package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"GoSnip/config"
	"GoSnip/db"
	"GoSnip/routes"
)

func main() {
	cfg := config.Load()

	pgDB := db.NewPostgres(cfg)
	defer pgDB.Close()

	redisClient := db.NewRedis(cfg)
	defer redisClient.Close()

	_ = redisClient // will be used from Phase 3 onwards

	r := gin.Default()

	routes.Setup(r, pgDB, cfg)

	log.Printf("GoSnip running on port %s", cfg.AppPort)

	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
