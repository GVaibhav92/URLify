package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type MetricsHandler struct {
	db *sqlx.DB
}

func NewMetricsHandler(db *sqlx.DB) *MetricsHandler {
	return &MetricsHandler{db: db}
}

type systemStats struct {
	TotalURLs   int `json:"total_urls"`
	ActiveURLs  int `json:"active_urls"`
	DownURLs    int `json:"down_urls"`
	UnknownURLs int `json:"unknown_urls"`
	TotalClicks int `json:"total_clicks"`
	TotalUsers  int `json:"total_users"`
}

func (h *MetricsHandler) GetStats(c *gin.Context) {
	var stats systemStats

	//aggregation in one round trip
	query := `
		SELECT
			COUNT(u.id)                                         AS total_urls,
			COUNT(CASE WHEN m.status = 'UP'      THEN 1 END)   AS active_urls,
			COUNT(CASE WHEN m.status = 'DOWN'    THEN 1 END)   AS down_urls,
			COUNT(CASE WHEN m.status = 'UNKNOWN' THEN 1 END)   AS unknown_urls,
			COALESCE(SUM(m.clicks), 0)                         AS total_clicks
		FROM urls u
		LEFT JOIN url_metrics m ON m.url_id = u.id
	`

	err := h.db.QueryRowx(query).Scan(
		&stats.TotalURLs,
		&stats.ActiveURLs,
		&stats.DownURLs,
		&stats.UnknownURLs,
		&stats.TotalClicks,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	// Separate count for users
	if err := h.db.QueryRowx(`SELECT COUNT(*) FROM users`).
		Scan(&stats.TotalUsers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user count"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
