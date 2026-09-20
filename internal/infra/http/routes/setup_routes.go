package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/uricampos/go-ticket-queue/internal/infra/http/events"
	"github.com/uricampos/go-ticket-queue/internal/infra/http/orders"
	"github.com/uricampos/go-ticket-queue/internal/infra/http/users"
	"github.com/uricampos/go-ticket-queue/internal/middlewares"
)

func SetupUsersRoutes(router *gin.Engine, userHandler *users.UserHandler, rateLimitMiddleware *middlewares.RateLimitMiddleware) {
	userGroup := router.Group("/users")
	userGroup.Use(rateLimitMiddleware.GetRateLimit())
	userGroup.POST("", userHandler.CreateUser)
	userGroup.GET("/:id", userHandler.GetUserByID)
	userGroup.GET("", userHandler.GetUserByUsername)
}

func SetupEventsRoutes(router *gin.Engine, eventHandler *events.EventHandler, rateLimitMiddleware *middlewares.RateLimitMiddleware) {
	eventGroup := router.Group("/events")
	eventGroup.Use(rateLimitMiddleware.GetRateLimit())
	eventGroup.POST("", eventHandler.CreateEvent)
	eventGroup.GET("/:id", eventHandler.GetEventByID)
}

func SetupOrderRoutes(router *gin.Engine, orderHandler *orders.OrderHandler, rateLimitMiddleware *middlewares.RateLimitMiddleware) {
	orderGroup := router.Group("/orders")
	orderGroup.Use(rateLimitMiddleware.GetRateLimit())
	orderGroup.POST("", orderHandler.CreateOrder)
	orderGroup.GET("/processed", orderHandler.GetOrdersProcessedCount)
	orderGroup.GET("/queue-size", orderHandler.GetQueueSize)
}
