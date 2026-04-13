package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	//HTTP Traffic
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "urlify_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Request duration- histogram
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "urlify_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Redirects
	RedirectCacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "urlify_redirect_cache_hits_total",
		Help: "Total redirect requests served from Redis cache",
	})

	RedirectCacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "urlify_redirect_cache_misses_total",
		Help: "Total redirect requests that required a PostgreSQL lookup",
	})

	RedirectsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "urlify_redirects_total",
		Help: "Total successful redirects performed",
	})

	//Rate Limiting
	RateLimitedRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "urlify_rate_limited_requests_total",
		Help: "Total requests rejected by rate limiter",
	})

	//URL Management
	ActiveURLsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "urlify_active_urls",
		Help: "Current number of registered URLs",
	})

	URLsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "urlify_urls_created_total",
		Help: "Total URLs created since startup",
	})

	URLsDeletedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "urlify_urls_deleted_total",
		Help: "Total URLs deleted since startup",
	})

	//Health Checker
	URLStatusGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "urlify_url_status",
			Help: "Current URL health status (1=UP, 0=DOWN)",
		},
		[]string{"short_code"},
	)

	HealthCheckCyclesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "urlify_health_check_cycles_total",
		Help: "Total health check cycles completed",
	})

	HealthCheckDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "urlify_health_check_duration_seconds",
		Help:    "Time taken to complete a full health check cycle",
		Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
	})
)
