package tickets

import (
	"sync"
)

type TicketStock struct {
	Type     string
	Quantity int
	BuysDone int
	sm       sync.Mutex
}

func (ts *TicketStock) Buy(quantity int, wg *sync.WaitGroup) *TicketStock {
	defer wg.Done()

	canBuyTicket := false

	ts.sm.Lock()

	if ts.Quantity > 0 && ts.Quantity >= quantity {
		canBuyTicket = true
		ts.BuysDone++
	}

	// time.Sleep(1 * time.Millisecond)

	if canBuyTicket {
		ts.Quantity -= quantity
	}

	ts.sm.Unlock()

	return ts
}
