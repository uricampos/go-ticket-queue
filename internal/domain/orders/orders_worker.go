package orders

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (s *OrderService) StartWorkers(i int) {
	for idx := range i {
		go func() {
			for job := range s.jobs {
				s.queueSize.Add(-1)
				slog.Info("order being processed by worker", "worker_id", idx, "user_id", job.userID, "idempotency_key", job.idempotencyKey)
				order, err := s.ProcessOrder(job.ctx, job.userID, job.idempotencyKey, job.totalPrice)
				job.result <- orderJobResult{order, err}
			}
		}()
	}
}

func (s *OrderService) ProcessOrder(ctx context.Context, userID uuid.UUID, idempotencyKey string, totalPrice decimal.Decimal) (*Order, error) {
	slog.Info("start processing order", "user_id", userID, "idempotency_key", idempotencyKey)
	order, err := s.repo.CreateOrder(ctx, userID, idempotencyKey, totalPrice)

	if err != nil {
		return nil, err
	}

	if order == nil {
		order, err = s.GetOrderByUserIDAndIdempotencyKey(ctx, userID, idempotencyKey)
		if err != nil {
			return nil, err
		}
		slog.Info("order already existed, idempotency conflict resolved", "user_id", userID, "idempotency_key", idempotencyKey, "order_id", order.ID, "status", order.Status)
		return order, nil
	}

	slog.Info("order created", "user_id", userID, "idempotency_key", idempotencyKey, "order_id", order.ID, "status", order.Status)
	return order, nil
}
