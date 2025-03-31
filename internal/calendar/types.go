package calendar

import (
	"context"
	"time"
)

// TimeSlot represents a time slot with start and end times
type TimeSlot struct {
	Start time.Time
	End   time.Time
}

// Event represents a calendar event with basic information
type Event struct {
	Summary        string
	Start          time.Time
	End            time.Time
	Description    string
	Attendees      []string
	OrganizerEmail string
}

// CalendarService defines the interface for calendar operations
type CalendarService interface {
	GetFreeSlots(ctx context.Context, email string, startTime, endTime time.Time, events []*Event) ([]TimeSlot, error)
	CreateEvent(ctx context.Context, email string, event *Event) error
}
