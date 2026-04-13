package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"URLify/config"
	"URLify/metrics"
)

//Lua script implements atomic token bucket refill + consume.

// KEYS[1] = the rate limit key
// ARGV[1] = max capacity (tokens)
// ARGV[2] = refill rate (tokens per second)
// ARGV[3] = current unix timestamp (seconds)

// Returns: { remaining_tokens, allowed }

const luaTokenBucket = `
local key          = KEYS[1]
local capacity     = tonumber(ARGV[1])
local refill_rate  = tonumber(ARGV[2])
local now          = tonumber(ARGV[3])
local cost         = 1

-- Load existing bucket state
local bucket = redis.call("HMGET", key, "tokens", "last_refill")
local tokens      = tonumber(bucket[1])
local last_refill = tonumber(bucket[2])

-- First request -> initialise bucket at full capacity
if tokens == nil then
    tokens      = capacity
    last_refill = now
end

-- how many tokens have refilled since last request
local elapsed        = math.max(0, now - last_refill)
local refilled       = math.floor(elapsed * refill_rate)
tokens               = math.min(capacity, tokens + refilled)
last_refill          = now

-- Attempt to consume one token
local allowed = 0
if tokens >= cost then
    tokens  = tokens - cost
    allowed = 1
end

-- Persist updated bucket state
-- TTL = capacity/refill_rate * 2  so idle buckets expire automatically
local ttl = math.ceil((capacity / refill_rate) * 2)
redis.call("HMSET", key, "tokens", tokens, "last_refill", last_refill)
redis.call("EXPIRE", key, ttl)

return { tokens, allowed }
`

var tokenBucketScript = redis.NewScript(luaTokenBucket)

func RateLimiter(rdb *redis.Client, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "ratelimit:" + ip

		now := time.Now().Unix()

		results, err := tokenBucketScript.Run(
			context.Background(),
			rdb,
			[]string{key},
			cfg.RateLimitCapacity,
			cfg.RateLimitRefillRate,
			now,
		).Slice()

		if err != nil {
			// fail open
			c.Next()
			return
		}

		remaining := results[0].(int64)
		allowed := results[1].(int64)

		//Expose standard rate limit headers
		c.Header("X-Ratelimit-Limit", strconv.Itoa(cfg.RateLimitCapacity))
		c.Header("X-Ratelimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-Ratelimit-Refill-Rate", strconv.Itoa(cfg.RateLimitRefillRate))

		if allowed == 0 {
			metrics.RateLimitedRequests.Inc()

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":               "Rate limit exceeded",
				"retry_after_seconds": cfg.RateLimitCapacity / cfg.RateLimitRefillRate,
			})
			return
		}

		c.Next()
	}
}
