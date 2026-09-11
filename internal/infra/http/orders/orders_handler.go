package orders

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

type OrderRequest struct {
	UserID     uuid.UUID       `json:"user_id" binding:"required"`
	TotalPrice decimal.Decimal `json:"total_price" binding:"required"`
}

type OrderResponse struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Status         string
	IdempotencyKey string
	TotalPrice     decimal.Decimal
	CreatedAt      time.Time
}

func (h *OrderHandler) CreateOrder(ctx *gin.Context) {
	var orderRequest OrderRequest

	err := ctx.ShouldBindJSON(&orderRequest)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	idempotencyKey := ctx.GetHeader("Idempotency-Key")

	if idempotencyKey == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "missing idempotency key",
		})
		return
	}

	order, err := h.svc.CreateOrder(ctx.Request.Context(), orderRequest.UserID, idempotencyKey, orderRequest.TotalPrice)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, OrderToResponse(*order))
}

func OrderToResponse(order orders.Order) *OrderResponse {
	return &OrderResponse{
		ID:             order.ID,
		UserID:         order.UserID,
		Status:         order.Status,
		IdempotencyKey: order.IdempotencyKey,
		TotalPrice:     order.TotalPrice,
		CreatedAt:      order.CreatedAt,
	}
}
