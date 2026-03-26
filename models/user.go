package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID        uuid.UUID `db:"id"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
}

type UserStore struct {
	db *sqlx.DB
}

func NewUserStore(db *sqlx.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(email, hashedPassword string) (*User, error) {
	user := &User{}

	query := `
		INSERT INTO users (email, password)
		VALUES ($1, $2)
		RETURNING id, email, password, role, created_at
	`

	err := s.db.QueryRowx(query, email, hashedPassword).StructScan(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserStore) GetByEmail(email string) (*User, error) {
	user := &User{}

	query := `
		SELECT id, email, password, role, created_at
		FROM users
		WHERE email = $1
	`

	err := s.db.QueryRowx(query, email).StructScan(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
