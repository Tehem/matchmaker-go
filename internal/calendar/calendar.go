package calendar

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// TimeSlot represents a time slot
type TimeSlot struct {
	Start time.Time
	End   time.Time
}

// Event represents a calendar event
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

// Service represents a Google Calendar service
type Service struct {
	service *calendar.Service
}

// calendarServiceFactory is a function that creates a new calendar service
var calendarServiceFactory = func(ctx context.Context, opts ...option.ClientOption) (*Service, error) {
	service, err := calendar.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &Service{service: service}, nil
}

// NewCalendarService creates a new calendar service
func NewCalendarService(ctx context.Context, opts ...option.ClientOption) (*Service, error) {
	return calendarServiceFactory(ctx, opts...)
}

// SetCalendarServiceFactory sets the factory function for creating calendar services
func SetCalendarServiceFactory(factory func(context.Context, ...option.ClientOption) (*Service, error)) {
	calendarServiceFactory = factory
}

// GetFreeSlots retrieves free slots for a given calendar ID and time range
func (s *Service) GetFreeSlots(ctx context.Context, email string, startTime, endTime time.Time, events []*Event) ([]TimeSlot, error) {
	if endTime.Before(startTime) {
		return nil, fmt.Errorf("end time %v is before start time %v", endTime, startTime)
	}

	// Convert calendar events to TimeSlots
	busySlots := make([]TimeSlot, 0, len(events))
	for _, event := range events {
		busySlots = append(busySlots, TimeSlot{
			Start: event.Start,
			End:   event.End,
		})
	}

	// Get free slots between busy slots
	freeSlots := calculateFreeSlots(startTime, endTime, busySlots)

	return freeSlots, nil
}

// CreateEvent creates a new calendar event
func (s *Service) CreateEvent(ctx context.Context, email string, event *Event) error {
	calendarEvent, err := s.transformEvent(event)
	if err != nil {
		return err
	}

	// If organizer email is set, add it as organizer and optional attendee
	if event.OrganizerEmail != "" {
		calendarEvent.Organizer = &calendar.EventOrganizer{
			Email: event.OrganizerEmail,
		}
		// Add organizer as optional attendee if not already in attendees list
		organizerExists := false
		for _, attendee := range calendarEvent.Attendees {
			if attendee.Email == event.OrganizerEmail {
				organizerExists = true
				break
			}
		}
		if !organizerExists {
			calendarEvent.Attendees = append(calendarEvent.Attendees, &calendar.EventAttendee{
				Email:    event.OrganizerEmail,
				Optional: true,
			})
		}
	}

	_, err = s.insertEvent(ctx, email, calendarEvent)
	return err
}

// transformEvent transforms an Event into a calendar.Event
func (s *Service) transformEvent(event *Event) (*calendar.Event, error) {
	calendarEvent := &calendar.Event{
		Summary:     event.Summary,
		Start:       &calendar.EventDateTime{DateTime: event.Start.Format(time.RFC3339)},
		End:         &calendar.EventDateTime{DateTime: event.End.Format(time.RFC3339)},
		Description: event.Description,
		Attendees:   make([]*calendar.EventAttendee, len(event.Attendees)),
	}

	for i, attendee := range event.Attendees {
		calendarEvent.Attendees[i] = &calendar.EventAttendee{
			Email: attendee,
		}
	}

	return calendarEvent, nil
}

// insertEvent inserts a calendar event into the calendar
func (s *Service) insertEvent(ctx context.Context, email string, event *calendar.Event) (*calendar.Event, error) {
	return s.service.Events.Insert(email, event).Context(ctx).Do()
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
