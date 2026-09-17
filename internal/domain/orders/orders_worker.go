package orders

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (s *OrderService) StartWorkers(i int) {
	for range i {
		go func() {
			for job := range s.jobs {
				order, err := s.ProcessOrder(job.ctx, job.userID, job.idempotencyKey, job.totalPrice)
				job.result <- orderJobResult{order, err}
			}
		}()
	}
}

func (s *OrderService) ProcessOrder(ctx context.Context, userID uuid.UUID, idempotencyKey string, totalPrice decimal.Decimal) (*Order, error) {
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
