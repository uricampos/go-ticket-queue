package events

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/uricampos/go-ticket-queue/internal/domain/events"
)

type EventsRepository struct {
	db *sql.DB
}

func NewEventsRepository(db *sql.DB) *EventsRepository {
	return &EventsRepository{
		db: db,
	}
}

func (r *EventsRepository) CreateEvent(ctx context.Context, name string, description string, category string, date time.Time, location string) (*events.Event, error) {
	var id uuid.UUID
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, "INSERT INTO events (name, description, category, date, location) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at",
		name, description, category, date, location).
		Scan(&id, &createdAt)

	if err != nil {
		return nil, err
	}

	return &events.Event{
		ID:          id,
		Name:        name,
		Description: description,
		Category:    category,
		Date:        date,
		Location:    location,
		CreatedAt:   createdAt,
	}, nil
}
