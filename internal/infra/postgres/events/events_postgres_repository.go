package events

import (
	"context"
	"database/sql"
	"errors"
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

func (r *EventsRepository) GetEventByID(ctx context.Context, id uuid.UUID) (*events.Event, error) {
	var name string
	var description string
	var category string
	var date time.Time
	var location string
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, "SELECT name, description, category, date, location, created_at FROM events WHERE id = $1", id).Scan(&name, &description, &category, &date, &location, &createdAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

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
