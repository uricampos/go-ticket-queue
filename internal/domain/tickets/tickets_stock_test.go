package tickets

import (
	"sync"
	"testing"
)

func TestTicketsStock_RaceCondition(t *testing.T) {

	var wg sync.WaitGroup

	ticketStock := &TicketStock{
		Type:     "full",
		Quantity: 1,
		BuysDone: 0,
	}

	for range 50 {
		wg.Add(1)
		go ticketStock.Buy(1, &wg)
	}

	wg.Wait()

	if ticketStock.Quantity < 0 || ticketStock.BuysDone > 1 {
		t.Fatal("Race Condition detected!")
	}
}

func BenchmarkTicketStock_Buy(b *testing.B) {
	ticketStock := &TicketStock{
		Type:     "full",
		Quantity: b.N,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var wg sync.WaitGroup
			wg.Add(1)
			ticketStock.Buy(1, &wg)
			wg.Wait()
		}
	})
}
