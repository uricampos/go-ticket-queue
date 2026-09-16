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

type buyRequest struct {
	quantity int
	result   chan bool
}

func (ts *TicketStock) StartWorker(request <-chan buyRequest) {
	go func() {
		for req := range request {
			approved := false
			if ts.Quantity > 0 && ts.Quantity >= req.quantity {
				ts.Quantity -= req.quantity
				approved = true
				ts.BuysDone++
			}
			req.result <- approved
		}
	}()
}
