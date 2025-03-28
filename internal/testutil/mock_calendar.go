package testutil

import (
	"context"
	"time"

	"matchmaker/internal/calendar"
)

// MockCalendarService implements a mock calendar service for testing
type MockCalendarService struct {
	CreatedEvents []struct {
		Email string
		Event *calendar.Event
	}
}

func (m *MockCalendarService) GetFreeSlots(ctx context.Context, email string, startTime, endTime time.Time, events []*calendar.Event) ([]calendar.TimeSlot, error) {
	return []calendar.TimeSlot{
		{
			Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
		},
	}, nil
}

func (m *MockCalendarService) CreateEvent(ctx context.Context, email string, event *calendar.Event) error {
	m.CreatedEvents = append(m.CreatedEvents, struct {
		Email string
		Event *calendar.Event
	}{
		Email: email,
		Event: event,
	})
	return nil
}

// Ensure MockCalendarService implements CalendarService
var _ calendar.CalendarService = &MockCalendarService{}
