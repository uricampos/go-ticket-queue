package users

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uricampos/go-ticket-queue/internal/domain/users"
)

type UserHandler struct {
	svc users.UserService
}

func NewUserHandler(svc users.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

type UserRequest struct {
	Username string `json:"username" binding:"required"`
}

type UserResponse struct {
	ID        uuid.UUID
	Username  string
	CreatedAt time.Time
}

func (h *UserHandler) CreateUser(ctx *gin.Context) {

	var userRequest UserRequest

	err := ctx.ShouldBindJSON(&userRequest)

	if userRequest.Username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "username must not be empty",
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.svc.CreateUser(ctx.Request.Context(), userRequest.Username)

	if errors.Is(err, users.ErrUsernameTaken) {
		ctx.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	userResponse := UserToResponse(*user)

	ctx.JSON(http.StatusCreated, userResponse)
}

func (h *UserHandler) GetUserByID(ctx *gin.Context) {
	id := ctx.Param("id")

	parsedID, err := uuid.Parse(id)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.svc.GetUserByID(ctx.Request.Context(), parsedID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	userResponse := UserToResponse(*user)

	ctx.JSON(http.StatusOK, userResponse)
}

func (h *UserHandler) GetUserByUsername(ctx *gin.Context) {
	username := ctx.Query("username")

	user, err := h.svc.GetUserByUsername(ctx.Request.Context(), username)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}
}

func UserToResponse(user users.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}
}
