package middleware

import (
	"URLify/metrics"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath() //returns route pattern /r/:shortcode

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method

		//prometheus metrics
		statusStr := strconv.Itoa(status)

		metrics.HTTPRequestsTotal.WithLabelValues(
			method,
			path,
			statusStr,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			method,
			path,
		).Observe(duration.Seconds())

		log.Printf(
			"[%s] %s %s | %d | %s | %s",
			method,
			path,
			c.Request.URL.RequestURI(),
			status,
			duration,
			c.ClientIP(),
		)

		if duration > 500*time.Millisecond {
			log.Printf("[SLOW REQUEST] %s %s took %s",
				method, path, duration)
		}
		if status >= 500 {
			errs := c.Errors.String()
			if errs != "" {
				log.Printf("[ERROR] %s %s | %d | %s",
					method, path, status, errs)
			}
		}
	}
}
