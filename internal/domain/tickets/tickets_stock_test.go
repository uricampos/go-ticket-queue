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
