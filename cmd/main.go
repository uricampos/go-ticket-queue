package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/uricampos/go-ticket-queue/internal/config"
	eventsDomain "github.com/uricampos/go-ticket-queue/internal/domain/events"
	usersDomain "github.com/uricampos/go-ticket-queue/internal/domain/users"
	eventsHandler "github.com/uricampos/go-ticket-queue/internal/infra/http/events"
	"github.com/uricampos/go-ticket-queue/internal/infra/http/routes"
	usersInfra "github.com/uricampos/go-ticket-queue/internal/infra/http/users"
	"github.com/uricampos/go-ticket-queue/internal/infra/postgres"
	eventsRepo "github.com/uricampos/go-ticket-queue/internal/infra/postgres/events"
	usersRepo "github.com/uricampos/go-ticket-queue/internal/infra/postgres/users"
)

func main() {
	// spin up env

	// config env
	cfg, err := config.Load()

	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.Connect(*cfg)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// setup routes
	router := gin.Default()

	// users
	userRepo := usersRepo.NewUserRepository(db)
	userService := usersDomain.NewUserService(userRepo)
	userHandler := usersInfra.NewUserHandler(*userService)

	routes.SetupUsersRoutes(router, userHandler)

	// events
	eventRepo := eventsRepo.NewEventsRepository(db)
	eventService := eventsDomain.NewEventService(eventRepo)
	eventHandler := eventsHandler.NewEventHandler(*eventService)

	routes.SetupEventsRoutes(router, eventHandler)

	// server listen
	router.Run(":" + cfg.ServerPort)
}
