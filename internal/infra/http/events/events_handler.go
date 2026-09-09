package events

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uricampos/go-ticket-queue/internal/domain/events"
)

type EventHandler struct {
	svc events.EventService
}

func NewEventHandler(svc events.EventService) *EventHandler {
	return &EventHandler{
		svc: svc,
	}
}

type EventRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Category    string    `json:"category" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	Location    string    `json:"location" binding:"required"`
}

type EventResponse struct {
	ID          uuid.UUID
	Name        string
	Description string
	Category    string
	Date        time.Time
	Location    string
	CreatedAt   time.Time
}

func (h *EventHandler) CreateEvent(ctx *gin.Context) {
	var eventRequest EventRequest

	err := ctx.ShouldBindJSON(&eventRequest)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	event, err := h.svc.CreateEvent(ctx.Request.Context(), eventRequest.Name, eventRequest.Description, eventRequest.Category, eventRequest.Date, eventRequest.Location)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, eventToResponse(*event))

}

func eventToResponse(event events.Event) EventResponse {
	return EventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		Category:    event.Category,
		Date:        event.Date,
		Location:    event.Location,
		CreatedAt:   event.CreatedAt,
	}
}
