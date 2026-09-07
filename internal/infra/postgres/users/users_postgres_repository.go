package users

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/uricampos/go-ticket-queue/internal/domain/users"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, username string) (*users.User, error) {

	var id uuid.UUID
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, "INSERT INTO users (username) VALUES ($1) RETURNING id, created_at", username).Scan(&id, &createdAt)

	if err != nil {
		return nil, err
	}

	return &users.User{
		ID:        id,
		Username:  username,
		CreatedAt: createdAt,
	}, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	var username string
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, "SELECT username, created_at FROM users WHERE id = $1", id).Scan(&username, &createdAt)
	if err != nil {
		return nil, err
	}

	return &users.User{
		ID:        id,
		Username:  username,
		CreatedAt: createdAt,
	}, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*users.User, error) {
	var id uuid.UUID
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, "SELECT id, created_at FROM users WHERE username = $1", username).Scan(&id, &createdAt)

	if err != nil {
		return nil, err
	}

	return &users.User{
		ID:        id,
		Username:  username,
		CreatedAt: createdAt,
	}, nil
}
