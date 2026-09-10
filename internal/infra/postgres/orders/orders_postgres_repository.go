package orders

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/uricampos/go-ticket-queue/internal/domain/orders"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, userID uuid.UUID, status string, idempotencyKey string, totalPrice decimal.Decimal) (*orders.Order, error) {
	var id uuid.UUID
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, "INSERT INTO orders (user_id, status, idempotency_key, total_price) VALUES ($1, $2, $3, $4) ON CONFLICT (user_id, idempotency_key) DO NOTHING RETURNING id, created_at", userID, status, idempotencyKey, totalPrice).Scan(&id, &createdAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &orders.Order{
		ID:             id,
		UserID:         userID,
		Status:         status,
		IdempotencyKey: idempotencyKey,
		TotalPrice:     totalPrice,
		CreatedAt:      createdAt,
	}, nil
}

func (r *OrderRepository) GetOrderByUserIDAndIdempotencyKey(ctx context.Context, userID uuid.UUID, idempotencyKey string) (*orders.Order, error) {
	var id uuid.UUID
	var status string
	var totalPrice decimal.Decimal
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, "SELECT id, status, total_price, created_at FROM orders WHERE user_id = $1 and idempotency_key = $2", userID, idempotencyKey).
		Scan(&id, &status, &totalPrice, &createdAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &orders.Order{
		ID:             id,
		Status:         status,
		UserID:         userID,
		IdempotencyKey: idempotencyKey,
		TotalPrice:     totalPrice,
		CreatedAt:      createdAt,
	}, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*orders.Order, error) {
	var status string
	var userID uuid.UUID
	var idempotencyKey string
	var totalPrice decimal.Decimal
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, "SELECT status, user_id, idempotency_key, total_price, created_at FROM orders WHERE id = $1", id).
		Scan(&status, &userID, &idempotencyKey, &totalPrice, &createdAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &orders.Order{
		ID:             id,
		UserID:         userID,
		Status:         status,
		IdempotencyKey: idempotencyKey,
		TotalPrice:     totalPrice,
		CreatedAt:      createdAt,
	}, nil
}
