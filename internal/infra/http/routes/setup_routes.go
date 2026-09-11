package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/uricampos/go-ticket-queue/internal/infra/http/events"
	"github.com/uricampos/go-ticket-queue/internal/infra/http/orders"
	"github.com/uricampos/go-ticket-queue/internal/infra/http/users"
)

func SetupUsersRoutes(router *gin.Engine, userHandler *users.UserHandler) {
	userGroup := router.Group("/users")
	userGroup.POST("", userHandler.CreateUser)
	userGroup.GET("/:id", userHandler.GetUserByID)
	userGroup.GET("", userHandler.GetUserByUsername)
}

func SetupEventsRoutes(router *gin.Engine, eventHandler *events.EventHandler) {
	eventGroup := router.Group("/events")
	eventGroup.POST("", eventHandler.CreateEvent)
	eventGroup.GET("/:id", eventHandler.GetEventByID)
}

func SetupOrderRoutes(router *gin.Engine, orderHandler *orders.OrderHandler) {
	orderGroup := router.Group("/orders")
	orderGroup.POST("", orderHandler.CreateOrder)
}
