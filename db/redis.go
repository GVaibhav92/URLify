package db

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	"URLify/config"
)

func NewRedis(cfg *config.Config) *redis.Client {
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	// Verify connection immediately
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis connection failed!: %v", err)
	}

	log.Println(" Redis connected")
	return client
}
