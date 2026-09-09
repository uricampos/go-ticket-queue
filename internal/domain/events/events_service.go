package events

import (
	"context"
	"time"
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
