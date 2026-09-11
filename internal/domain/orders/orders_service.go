package orders

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID, idempotencyKey string, totalPrice decimal.Decimal) (*Order, error) {
	order, err := s.repo.CreateOrder(ctx, userID, idempotencyKey, totalPrice)

	if err != nil {
		return nil, err
	}

	if order == nil {
		order, err = s.GetOrderByUserIDAndIdempotencyKey(ctx, userID, idempotencyKey)
		if err != nil {
			return nil, err
		}
	}

	return order, nil
}

func (s *OrderService) GetOrderByUserIDAndIdempotencyKey(ctx context.Context, userID uuid.UUID, idempotencyKey string) (*Order, error) {
	order, err := s.repo.GetOrderByUserIDAndIdempotencyKey(ctx, userID, idempotencyKey)

	if err != nil {
		return nil, err
	}

	return order, err
}

func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	order, err := s.repo.GetOrderByID(ctx, id)

	if err != nil {
		return nil, err
	}

	return order, err
}
