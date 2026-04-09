package services

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"URLify/models"
)

const (
	redirectKeyPrefix = "redirect:"
	redirectTTL       = 24 * time.Hour
)

type RedirectService struct {
	urlStore *models.URLStore
	redis    *redis.Client
}

func NewRedirectService(urlStore *models.URLStore, rdb *redis.Client) *RedirectService {
	return &RedirectService{
		urlStore: urlStore,
		redis:    rdb,
	}
}

func (s *RedirectService) Resolve(ctx context.Context, shortCode string) (string, bool, error) {
	key := redirectKeyPrefix + shortCode

	//check Redis
	val, err := s.redis.Get(ctx, key).Result()

	if err == nil {
		//cache HIT
		return val, true, nil
	}

	if !errors.Is(err, redis.Nil) {
		//redis.Nil = “key not found” - log but don't abort
		//fall through to PostgreSQL so redirect still works
	}

	// cache MISS
	url, err := s.urlStore.GetByShortCode(shortCode)
	if err != nil {
		return "", false, err // shortcode not found
	}

	//store in Redis
	s.redis.Set(ctx, key, url.OriginalURL, redirectTTL)

	return url.OriginalURL, false, nil
}

func (s *RedirectService) InvalidateCache(ctx context.Context, shortCode string) error {
	key := redirectKeyPrefix + shortCode
	return s.redis.Del(ctx, key).Err()
}
