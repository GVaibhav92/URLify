package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type URL struct {
	ID          uuid.UUID  `db:"id"`
	ShortCode   string     `db:"short_code"`
	OriginalURL string     `db:"original_url"`
	UserID      uuid.UUID  `db:"user_id"`
	IsCustom    bool       `db:"is_custom"`
	CreatedAt   time.Time  `db:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at"`
}

type URLMetrics struct {
	ID               uuid.UUID  `db:"id"`
	URLID            uuid.UUID  `db:"url_id"`
	Clicks           int        `db:"clicks"`
	LastChecked      *time.Time `db:"last_checked"`
	Status           string     `db:"status"`
	UptimePercentage float64    `db:"uptime_percentage"`
}

type URLWithMetrics struct {
	ID               uuid.UUID  `db:"id"                json:"id"`
	ShortCode        string     `db:"short_code"        json:"short_code"`
	OriginalURL      string     `db:"original_url"      json:"original_url"`
	IsCustom         bool       `db:"is_custom"         json:"is_custom"`
	CreatedAt        time.Time  `db:"created_at"        json:"created_at"`
	ExpiresAt        *time.Time `db:"expires_at"        json:"expires_at"`
	Clicks           int        `db:"clicks"            json:"clicks"`
	Status           string     `db:"status"            json:"status"`
	UptimePercentage float64    `db:"uptime_percentage" json:"uptime_percentage"`
}

type URLStore struct {
	db *sqlx.DB
}

func NewURLStore(db *sqlx.DB) *URLStore {
	return &URLStore{db: db}
}

//Queries

func (s *URLStore) Create(originalURL, shortCode string, userID uuid.UUID, isCustom bool) (*URL, error) {
	url := &URL{}

	query := `
		INSERT INTO urls (original_url, short_code, user_id, is_custom)
		VALUES ($1, $2, $3, $4)
		RETURNING id, original_url, short_code, user_id, is_custom, created_at, expires_at
	`

	err := s.db.QueryRowx(query, originalURL, shortCode, userID, isCustom).StructScan(url)
	if err != nil {
		return nil, err
	}

	metricsQuery := `
		INSERT INTO url_metrics (url_id)
		VALUES ($1)
	`
	if _, err := s.db.Exec(metricsQuery, url.ID); err != nil {
		return nil, err
	}

	return url, nil
}

func (s *URLStore) GetByUserID(userID uuid.UUID, limit, offset int) ([]URLWithMetrics, error) {
	urls := []URLWithMetrics{}

	query := `
		SELECT
			u.id, u.short_code, u.original_url, u.is_custom,
			u.created_at, u.expires_at,
			m.clicks, m.status, m.uptime_percentage
		FROM urls u
		JOIN url_metrics m ON m.url_id = u.id
		WHERE u.user_id = $1
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := s.db.Select(&urls, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	return urls, nil
}

func (s *URLStore) CountByUserID(userID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM urls WHERE user_id = $1`
	err := s.db.QueryRowx(query, userID).Scan(&count)
	return count, err
}

func (s *URLStore) GetByID(id uuid.UUID) (*URL, error) {
	url := &URL{}

	query := `
		SELECT id, original_url, short_code, user_id, is_custom, created_at, expires_at
		FROM urls
		WHERE id = $1
	`

	err := s.db.QueryRowx(query, id).StructScan(url)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (s *URLStore) GetByShortCode(shortCode string) (*URL, error) {
	url := &URL{}

	query := `
		SELECT id, original_url, short_code, user_id, is_custom, created_at, expires_at
		FROM urls
		WHERE short_code = $1
	`

	err := s.db.QueryRowx(query, shortCode).StructScan(url)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (s *URLStore) Delete(id uuid.UUID) error {
	query := `DELETE FROM urls WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *URLStore) ShortCodeExists(shortCode string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM urls WHERE short_code = $1`
	err := s.db.QueryRowx(query, shortCode).Scan(&count)
	return count > 0, err
}
