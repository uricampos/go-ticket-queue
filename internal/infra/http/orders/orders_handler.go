package orders

import (
	"github.com/gin-gonic/gin"
	"github.com/uricampos/go-ticket-queue/internal/domain/orders"
)

type OrderHandler struct {
	svc orders.OrderService
}

func NewOrderHandler(svc orders.OrderService) *OrderHandler {
	return &OrderHandler{
		svc: svc,
	}
}

func (h *OrderHandler) CreateOrder(ctx *gin.Context) {

}
