package orders

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderService struct {
	repo      OrderRepository
	jobs      chan orderJob
	queueSize atomic.Int64
}

type orderJobResult struct {
	order *Order
	err   error
}

type orderJob struct {
	ctx            context.Context
	userID         uuid.UUID
	idempotencyKey string
	totalPrice     decimal.Decimal
	result         chan orderJobResult
}

func NewOrderService(repo OrderRepository) *OrderService {
	s := &OrderService{
		repo: repo,
		jobs: make(chan orderJob),
	}

	s.StartWorkers(5)
	return s
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID, idempotencyKey string, totalPrice decimal.Decimal) (*Order, error) {
	resultChan := make(chan orderJobResult)
	s.queueSize.Add(1)
	s.jobs <- orderJob{ctx, userID, idempotencyKey, totalPrice, resultChan}
	slog.Info("order job enqueued", "user_id", userID, "idempotency_key", idempotencyKey)
	res := <-resultChan

	if res.err != nil {
		slog.Error("order processing failed", "user_id", userID, "idempotency_key", idempotencyKey, "error", res.err)
		return nil, res.err
	}

	slog.Info("order processed", "user_id", userID, "idempotency_key", idempotencyKey, "order_id", res.order.ID, "status", res.order.Status)
	return res.order, nil
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

func (s *OrderService) GetOrdersProcessedCount(ctx context.Context) (int, error) {
	return s.repo.GetOrdersProcessedCount(ctx)
}

func (s *OrderService) GetQueueSize() int {
	return int(s.queueSize.Load())
}
