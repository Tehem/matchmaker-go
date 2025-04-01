package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"os/user"
	"path/filepath"
	"time"

	"matchmaker/internal/fs"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Service represents a Google Calendar service
type Service struct {
	service *calendar.Service
}

// NewCalendarServiceFromToken creates a new calendar service using the token file
func NewCalendarServiceFromToken(ctx context.Context) (*Service, error) {
	// Get user's home directory
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("unable to get user's home directory: %w", err)
	}
	return newCalendarServiceFromTokenWithHome(ctx, usr.HomeDir, fs.Default)
}

// newCalendarServiceFromTokenWithHome is an internal function that accepts a home directory for testing
func newCalendarServiceFromTokenWithHome(ctx context.Context, homeDir string, filesystem fs.FileSystem) (*Service, error) {
	// Get token file path
	tokenFile := filepath.Join(homeDir, ".credentials", "calendar-api.json")

	// Read token file
	data, err := filesystem.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	// Decode token
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	// Create OAuth2 client
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&token))

	// Create calendar service using the common implementation
	return NewCalendarService(ctx, option.WithHTTPClient(client))
}

// NewCalendarService creates a new calendar service with custom options
func NewCalendarService(ctx context.Context, opts ...option.ClientOption) (*Service, error) {
	service, err := calendar.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &Service{service: service}, nil
}

// GetFreeSlots implements CalendarService.GetFreeSlots
func (s *Service) GetFreeSlots(ctx context.Context, email string, startTime, endTime time.Time, events []*Event) ([]TimeSlot, error) {
	// Convert events to time slots for easier processing
	busySlots := make([]TimeSlot, len(events))
	for i, event := range events {
		busySlots[i] = TimeSlot{Start: event.Start, End: event.End}
	}

	return GetFreeSlots(startTime, endTime, busySlots)
}

// CreateEvent implements CalendarService.CreateEvent
func (s *Service) CreateEvent(ctx context.Context, email string, event *Event) error {
	calendarEvent, err := s.transformEvent(event)
	if err != nil {
		return err
	}

	_, err = s.insertEvent(ctx, email, calendarEvent)
	return err
}

// transformEvent converts our Event type to Google Calendar's Event type
func (s *Service) transformEvent(event *Event) (*calendar.Event, error) {
	calendarEvent := &calendar.Event{
		Summary:     event.Summary,
		Start:       &calendar.EventDateTime{DateTime: event.Start.Format(time.RFC3339)},
		End:         &calendar.EventDateTime{DateTime: event.End.Format(time.RFC3339)},
		Description: event.Description,
		Attendees:   make([]*calendar.EventAttendee, 0, len(event.Attendees)+1), // Pre-allocate with capacity for potential organizer
	}

	// Add regular attendees
	for _, attendee := range event.Attendees {
		calendarEvent.Attendees = append(calendarEvent.Attendees, &calendar.EventAttendee{
			Email: attendee,
		})
	}

	// Set organizer if provided
	if event.OrganizerEmail != "" {
		calendarEvent.Organizer = &calendar.EventOrganizer{
			Email: event.OrganizerEmail,
		}
		// Add organizer as optional attendee if not already present
		addOrganizerAsAttendee(calendarEvent, event.OrganizerEmail)
	}

	return calendarEvent, nil
}

// addOrganizerAsAttendee adds the organizer as an optional attendee if not already present
func addOrganizerAsAttendee(event *calendar.Event, organizerEmail string) {
	for _, attendee := range event.Attendees {
		if attendee.Email == organizerEmail {
			return
		}
	}
	event.Attendees = append(event.Attendees, &calendar.EventAttendee{
		Email:    organizerEmail,
		Optional: true,
	})
}

// insertEvent inserts a calendar event into Google Calendar
func (s *Service) insertEvent(ctx context.Context, email string, event *calendar.Event) (*calendar.Event, error) {
	return s.service.Events.Insert(email, event).Context(ctx).Do()
}

// GetBusySlots retrieves busy time slots for a given email and time range
func (s *Service) GetBusySlots(ctx context.Context, email string, startTime, endTime time.Time) ([]TimeSlot, error) {
	// Create FreeBusy request
	freeBusyRequest := &calendar.FreeBusyRequest{
		TimeMin: startTime.Format(time.RFC3339),
		TimeMax: endTime.Format(time.RFC3339),
		Items: []*calendar.FreeBusyRequestItem{
			{Id: email},
		},
	}

	// Execute FreeBusy query
	freeBusyResponse, err := s.service.Freebusy.Query(freeBusyRequest).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query free/busy: %w", err)
	}

	// Extract busy slots from response
	busySlots := make([]TimeSlot, 0)
	if calendarBusy, ok := freeBusyResponse.Calendars[email]; ok {
		for _, busy := range calendarBusy.Busy {
			start, err := time.Parse(time.RFC3339, busy.Start)
			if err != nil {
				return nil, fmt.Errorf("failed to parse busy start time: %w", err)
			}
			end, err := time.Parse(time.RFC3339, busy.End)
			if err != nil {
				return nil, fmt.Errorf("failed to parse busy end time: %w", err)
			}
			busySlots = append(busySlots, TimeSlot{Start: start, End: end})
		}
	}

	return busySlots, nil
}
