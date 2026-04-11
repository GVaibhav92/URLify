package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"URLify/config"
	"URLify/db"
	"URLify/routes"
	"URLify/worker"
)

func main() {
	cfg := config.Load()

	pgDB := db.NewPostgres(cfg)
	defer pgDB.Close()

	redisClient := db.NewRedis(cfg)
	defer redisClient.Close()

	//context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//background health checker worker
	checker := worker.NewHealthChecker(pgDB, cfg)
	go func() {
		checker.Start(ctx)
	}()

	//Gin router
	r := gin.Default()
	routes.Setup(r, pgDB, redisClient, cfg)

	//HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	go func() {
		log.Printf("URLify running on port %s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	//block until shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown signal recieved...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("URLify shutdown complete")
}
