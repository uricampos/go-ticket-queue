package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/uricampos/go-ticket-queue/internal/infra/http/users"
)

func SetupUsersRoutes(router *gin.Engine, userHandler *users.UserHandler) {
	userGroup := router.Group("/users")
	userGroup.POST("", userHandler.CreateUser)
	userGroup.GET("/:id", userHandler.GetUserByID)
	userGroup.GET("", userHandler.GetUserByUsername)
}
