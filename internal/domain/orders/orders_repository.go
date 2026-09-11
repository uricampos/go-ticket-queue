package orders

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Order struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Status         string
	IdempotencyKey string
	TotalPrice     decimal.Decimal
	CreatedAt      time.Time
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, idempotencyKey string, totalPrice decimal.Decimal) (*Order, error)
	GetOrderByUserIDAndIdempotencyKey(ctx context.Context, userID uuid.UUID, idempotencyKey string) (*Order, error)
	GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error)
}
