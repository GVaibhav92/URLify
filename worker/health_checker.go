package worker

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"URLify/config"
	"URLify/metrics"
	"URLify/models"
)

const (
	StatusUP      = "UP"
	StatusDOWN    = "DOWN"
	StatusTimeout = "TIMEOUT"
	StatusUnknown = "UNKNOWN"
)

// Job-represents a URL to check
type job struct {
	urlID       string
	originalURL string
	shortCode   string
}

// HealthChecker
type HealthChecker struct {
	db         *sqlx.DB
	urlStore   *models.URLStore
	cfg        *config.Config
	httpClient *http.Client
}

func NewHealthChecker(db *sqlx.DB, cfg *config.Config) *HealthChecker {
	return &HealthChecker{
		db:       db,
		urlStore: models.NewURLStore(db),
		cfg:      cfg,
		// HTTP client
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			// Don't follow redirects-> check only that the URL is alive
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// Start launches the ticker loop and worker pool.
// blocks until ctx is cancelled
func (h *HealthChecker) Start(ctx context.Context) {
	interval := time.Duration(h.cfg.HealthCheckIntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf(" Health checker started (interval: %s, workers: %d)",
		interval, h.cfg.HealthCheckWorkerPoolSize)

	// Run one check immediately on startup to not for
	// the full interval before seeing any health data
	h.runCycle(ctx)

	for {
		select {
		//periodic execution of health checks
		case <-ticker.C:
			h.runCycle(ctx)

		case <-ctx.Done():
			log.Println("Health checker shutting down")
			return
		}
	}
}

// runCycle fetches all URLs and dispatches them to the worker pool via channel
func (h *HealthChecker) runCycle(ctx context.Context) {
	cycleStart := time.Now()
	defer func() {
		metrics.HealthCheckCyclesTotal.Inc()
		metrics.HealthCheckDuration.Observe(time.Since(cycleStart).Seconds())

	}()

	urls, err := h.urlStore.GetAllURLs()
	if err != nil {
		log.Printf("Health checker: failed to fetch URLs: %v", err)
		return
	}

	if len(urls) == 0 {
		return
	}

	log.Printf("Health checker: checking %d URLs", len(urls))

	//buffered channel so all jobs can be queued instantly without blocking.
	jobs := make(chan job, len(urls))

	// waits until all workers finish to prevent overlapping cycles.
	var wg sync.WaitGroup

	// Launch fixed pool of N workers
	for i := 0; i < h.cfg.HealthCheckWorkerPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.worker(ctx, jobs)
		}()
	}

	// Push all jobs into the channel
	for _, u := range urls {
		jobs <- job{
			urlID:       u.ID.String(),
			originalURL: u.OriginalURL,
			shortCode:   u.ShortCode,
		}
	}

	// close channel , signals workers that no more jobs are coming
	// workers exit their range loop and return, wg.Done() fires
	close(jobs)

	// wait for all workers to finish before returning
	wg.Wait()

	log.Printf("Health checker: cycle complete for %d URLs", len(urls))
}

// worker pulls jobs from the channel,processes each URL and exits cleanly
func (h *HealthChecker) worker(ctx context.Context, jobs <-chan job) {
	for j := range jobs {
		// Check if context was cancelled between jobs
		select {
		case <-ctx.Done():
			return
		default:
		}

		status := h.checkURL(ctx, j.originalURL)
		h.updateMetrics(j.urlID, j.shortCode, status)
	}
}

// checkURL performs the HTTP request using a context-aware client,
// and returns health status string
func (h *HealthChecker) checkURL(ctx context.Context, rawURL string) string {
	// Build request with context so it aligns with cancellation
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return StatusDOWN
	}

	// Identify ourselves politely
	req.Header.Set("User-Agent", "URLify-HealthChecker/1.0")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() != nil {
			return StatusTimeout
		}
		return StatusDOWN
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return StatusUP
	}

	return StatusDOWN
}

// updateMetrics writes the check result to url_metrics and adjusts the uptime percentage
func (h *HealthChecker) updateMetrics(urlID, shortCode, status string) {
	query := `
		UPDATE url_metrics
		SET
			status       = $1,
			last_checked = NOW(),
			uptime_percentage = CASE
				WHEN $1 = 'UP'
				THEN LEAST(100.0, uptime_percentage + 1.0)
				ELSE GREATEST(0.0, uptime_percentage - 2.0)
			END
		WHERE url_id = $2
	`

	if _, err := h.db.Exec(query, status, urlID); err != nil {
		log.Printf("Health checker: failed to update metrics for %s: %v", urlID, err)
	}

	//update prometheus gauge
	val := 0.0
	if status == StatusUP {
		val = 1.0
	}
	metrics.URLStatusGauge.WithLabelValues(shortCode).Set(val)
}
