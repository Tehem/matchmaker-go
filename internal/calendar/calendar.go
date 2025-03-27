package calendar

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Service represents a Google Calendar service
type Service struct {
	service *calendar.Service
}

// NewService creates a new calendar service
func NewService(ctx context.Context, credentialsFile string) (*Service, error) {
	service, err := calendar.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	return &Service{service: service}, nil
}

// GetFreeSlots retrieves free slots for a given calendar ID and time range
func (s *Service) GetFreeSlots(ctx context.Context, calendarID string, start, end time.Time) ([]TimeSlot, error) {
	events, err := s.service.Events.List(calendarID).
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// Convert events to busy slots
	busySlots := make([]TimeSlot, 0, len(events.Items))
	for _, event := range events.Items {
		startTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse event start time: %w", err)
		}

		endTime, err := time.Parse(time.RFC3339, event.End.DateTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse event end time: %w", err)
		}

		busySlots = append(busySlots, TimeSlot{
			Start: startTime,
			End:   endTime,
		})
	}

	// Calculate free slots
	return calculateFreeSlots(start, end, busySlots), nil
}

// CreateEvent creates a new calendar event
func (s *Service) CreateEvent(ctx context.Context, calendarID string, event *Event) error {
	calendarEvent := &calendar.Event{
		Summary: event.Summary,
		Start: &calendar.EventDateTime{
			DateTime: event.Start.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: event.End.Format(time.RFC3339),
		},
		Description: event.Description,
		Attendees:   make([]*calendar.EventAttendee, len(event.Attendees)),
	}

	for i, attendee := range event.Attendees {
		calendarEvent.Attendees[i] = &calendar.EventAttendee{
			Email: attendee,
		}
	}

	_, err := s.service.Events.Insert(calendarID, calendarEvent).
		SendUpdates("all").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// TimeSlot represents a time slot
type TimeSlot struct {
	Start time.Time
	End   time.Time
}

// Event represents a calendar event
type Event struct {
	Summary     string
	Start       time.Time
	End         time.Time
	Description string
	Attendees   []string
}

// calculateFreeSlots calculates free time slots between busy slots
func calculateFreeSlots(start, end time.Time, busySlots []TimeSlot) []TimeSlot {
	if len(busySlots) == 0 {
		return []TimeSlot{{Start: start, End: end}}
	}

	// Sort busy slots by start time
	sortTimeSlots(busySlots)

	freeSlots := make([]TimeSlot, 0)
	currentTime := start

	for _, busySlot := range busySlots {
		if currentTime.Before(busySlot.Start) {
			freeSlots = append(freeSlots, TimeSlot{
				Start: currentTime,
				End:   busySlot.Start,
			})
		}
		currentTime = busySlot.End
	}

	if currentTime.Before(end) {
		freeSlots = append(freeSlots, TimeSlot{
			Start: currentTime,
			End:   end,
		})
	}

	return freeSlots
}

// sortTimeSlots sorts time slots by start time
func sortTimeSlots(slots []TimeSlot) {
	// Implementation of a simple bubble sort
	for i := 0; i < len(slots)-1; i++ {
		for j := 0; j < len(slots)-i-1; j++ {
			if slots[j].Start.After(slots[j+1].Start) {
				slots[j], slots[j+1] = slots[j+1], slots[j]
			}
		}
	}
}
