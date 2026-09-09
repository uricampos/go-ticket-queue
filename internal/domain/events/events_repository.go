package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID
	Name        string
	Description string
	Category    string
	Date        time.Time
	Location    string
	CreatedAt   time.Time
}

type EventRepository interface {
	CreateEvent(ctx context.Context, name string, description string, category string, date time.Time, location string) (*Event, error)
	GetEventByID(ctx context.Context, id uuid.UUID) (*Event, error)
}
