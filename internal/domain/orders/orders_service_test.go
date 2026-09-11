package orders_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/uricampos/go-ticket-queue/internal/config"
	ordersDomain "github.com/uricampos/go-ticket-queue/internal/domain/orders"
	"github.com/uricampos/go-ticket-queue/internal/infra/postgres"
	ordersRepo "github.com/uricampos/go-ticket-queue/internal/infra/postgres/orders"
)

func TestCreateOrder_ConcurrentSameIdempotencyKey_CreatesOnlyOndeOrder(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	db, err := postgres.Connect(*cfg)
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	repo := ordersRepo.NewOrderRepository(db)
	service := ordersDomain.NewOrderService(repo)

	userID := "ef35b68f-9205-493d-92cf-518316dd88c7"
	parsedUserID, _ := uuid.Parse(userID)
	idempotencyKey := uuid.NewString()
	totalPrice := decimal.NewFromInt(100)

	const numRequest = 100

	var wg sync.WaitGroup
	results := make([]*ordersDomain.Order, numRequest)
	errs := make([]error, numRequest)

	for i := 0; i < numRequest; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			order, err := service.CreateOrder(context.Background(), parsedUserID, idempotencyKey, totalPrice)
			results[i] = order
			errs[i] = err
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		if results[i] == nil {
			t.Fatalf("request %d returned a nil order without an error", i)
		}
	}

	firstID := results[0].ID
	for i := range results {
		if results[i].ID != firstID {
			t.Errorf("request %d returned a different order ID: got %v, want %v", i, results[i].ID, firstID)
		}
	}

	var ordersResultCount int

	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM orders WHERE user_id = $1 AND idempotency_key = $2", parsedUserID, idempotencyKey).Scan(&ordersResultCount)

	if err != nil {
		t.Fatalf("error on counting orders: %v", err.Error())
	}

	if ordersResultCount != 1 {
		t.Errorf("expected 1 orders, got %d", ordersResultCount)
	}
}
