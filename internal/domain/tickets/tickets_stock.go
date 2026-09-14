package tickets

import (
	"sync"
	"time"
)

type TicketStock struct {
	Type     string
	Quantity int
	BuysDone int
}

func (ts *TicketStock) Buy(quantity int, wg *sync.WaitGroup) *TicketStock {
	defer wg.Done()

	canBuyTicket := false

	if ts.Quantity > 0 && ts.Quantity >= quantity {
		canBuyTicket = true
		ts.BuysDone++
	}

	time.Sleep(1 * time.Millisecond)

	if canBuyTicket {
		ts.Quantity -= quantity
	}

	return ts
}
