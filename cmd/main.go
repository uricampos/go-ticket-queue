package main

import (
	"log"

	"github.com/uricampos/go-ticket-queue/internal/config"
	"github.com/uricampos/go-ticket-queue/internal/infra/postgres"
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

	// server listen
}
