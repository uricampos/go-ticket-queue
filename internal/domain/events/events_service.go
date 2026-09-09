package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type EventService struct {
	repo EventRepository
}

func NewEventService(repo EventRepository) *EventService {
	return &EventService{
		repo: repo,
	}
}

func (s *EventService) CreateEvent(ctx context.Context, name string, description string, category string, date time.Time, location string) (*Event, error) {
	event, err := s.repo.CreateEvent(ctx, name, description, category, date, location)

	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) GetEventByID(ctx context.Context, id uuid.UUID) (*Event, error) {
	event, err := s.repo.GetEventByID(ctx, id)

	if err != nil {
		return nil, err
	}

	return event, nil
}
