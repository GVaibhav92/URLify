package db

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"GoSnip/config"
)

// dependency factory
func NewPostgres(cfg *config.Config) *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("PostgreSQL connection failed!: %v", err)
	}

	// connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	log.Println("PostgreSQL connected!")

	runMigrations(db)

	return db
}

func runMigrations(db *sqlx.DB) {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

	CREATE TABLE IF NOT EXISTS users (
		id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		email      TEXT UNIQUE NOT NULL,
		password   TEXT NOT NULL,
		role       TEXT NOT NULL DEFAULT 'free',
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS urls (
		id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		short_code   TEXT UNIQUE NOT NULL,
		original_url TEXT NOT NULL,
		user_id      UUID REFERENCES users(id) ON DELETE CASCADE,
		is_custom    BOOLEAN NOT NULL DEFAULT false,
		created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
		expires_at   TIMESTAMP NULL
	);

	CREATE TABLE IF NOT EXISTS url_metrics (
		id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		url_id            UUID UNIQUE REFERENCES urls(id) ON DELETE CASCADE,
		clicks            INTEGER NOT NULL DEFAULT 0,
		last_checked      TIMESTAMP,
		status            TEXT NOT NULL DEFAULT 'UNKNOWN',
		uptime_percentage FLOAT NOT NULL DEFAULT 0.0
	);
	`

	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations applied")
}
