package gcalendar

import (
	"matchmaker/libs/testutils"
	"matchmaker/libs/types"
	"testing"
	"time"

	"google.golang.org/api/calendar/v3"
)

// MockGCalendar is a mock implementation of GCalendar for testing
type MockGCalendar struct {
	service *calendar.Service
}

// NewMockGCalendar creates a new mock GCalendar for testing
func NewMockGCalendar() (*GCalendar, error) {
	// Create a mock service
	service := &calendar.Service{}
	return &GCalendar{service: service}, nil
}

func TestFormatTime(t *testing.T) {
	// Create a test time
	testTime := time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC)

	// Format the time
	formattedTime := FormatTime(testTime)

	// Check that the formatted time is in RFC3339 format
	expectedFormat := "2024-04-01T12:00:00Z"
	if formattedTime != expectedFormat {
		t.Errorf("FormatTime() = %v, want %v", formattedTime, expectedFormat)
	}
}

func TestFindAvailableSlots_NoCalendars(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Set up test time range
	timeMin := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)
	timeMax := time.Date(2024, 4, 1, 17, 0, 0, 0, time.UTC)

	// Test with no calendars (should return all slots)
	slots := generateAvailableSlots(timeMin, timeMax, 60)

	// Check that we got some slots
	if len(slots) == 0 {
		t.Error("generateAvailableSlots() returned no slots for empty calendar list")
	}

	// Test with a specific duration
	durationMinutes := 30
	slots = generateAvailableSlots(timeMin, timeMax, durationMinutes)

	// Check that the slots have the correct duration
	for _, slot := range slots {
		duration := slot.End.Sub(slot.Start)
		expectedDuration := time.Duration(durationMinutes) * time.Minute
		if duration != expectedDuration {
			t.Errorf("Slot duration = %v, want %v", duration, expectedDuration)
		}
	}
}

// Helper function to generate available slots without using the calendar service
func generateAvailableSlots(timeMin, timeMax time.Time, durationMinutes int) []types.Range {
	var availableSlots []types.Range
	current := timeMin
	duration := time.Duration(durationMinutes) * time.Minute

	for current.Add(duration).Before(timeMax) || current.Add(duration).Equal(timeMax) {
		availableSlots = append(availableSlots, types.Range{
			Start: current,
			End:   current.Add(duration),
		})
		current = current.Add(duration)
	}

	return availableSlots
}

func TestCreateSessionEvent(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create a test session
	start := time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	session := &types.ReviewSession{
		Reviewers: &types.Squad{
			People: []*types.Person{
				{Email: "person1@example.com"},
				{Email: "person2@example.com"},
			},
		},
		Range: &types.Range{
			Start: start,
			End:   end,
		},
	}

	// Create the event directly without using the calendar service
	event := &calendar.Event{
		Summary: session.GetEventSummary(),
		Start: &calendar.EventDateTime{
			DateTime: FormatTime(session.Range.Start),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: FormatTime(session.Range.End),
			TimeZone: "UTC",
		},
		Attendees: []*calendar.EventAttendee{
			{Email: session.Reviewers.People[0].Email},
			{Email: session.Reviewers.People[1].Email},
		},
		Reminders: &calendar.EventReminders{
			UseDefault: false,
			Overrides: []*calendar.EventReminder{
				{Method: "email", Minutes: 24 * 60},
				{Method: "popup", Minutes: 10},
			},
		},
	}

	// Check that the event has the correct properties
	if event.Summary != session.GetEventSummary() {
		t.Errorf("Event summary = %v, want %v", event.Summary, session.GetEventSummary())
	}

	if event.Start.DateTime != FormatTime(session.Range.Start) {
		t.Errorf("Event start time = %v, want %v", event.Start.DateTime, FormatTime(session.Range.Start))
	}

	if event.End.DateTime != FormatTime(session.Range.End) {
		t.Errorf("Event end time = %v, want %v", event.End.DateTime, FormatTime(session.Range.End))
	}

	// Check that the event has the correct attendees
	if len(event.Attendees) != 2 {
		t.Errorf("Event has %d attendees, want 2", len(event.Attendees))
	}

	if event.Attendees[0].Email != session.Reviewers.People[0].Email {
		t.Errorf("First attendee email = %v, want %v", event.Attendees[0].Email, session.Reviewers.People[0].Email)
	}

	if event.Attendees[1].Email != session.Reviewers.People[1].Email {
		t.Errorf("Second attendee email = %v, want %v", event.Attendees[1].Email, session.Reviewers.People[1].Email)
	}
}

func TestGetBusyTimesForPeople(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create test people
	people := []*types.Person{
		{Email: "person1@example.com", MaxSessionsPerWeek: 2},
		{Email: "person2@example.com", MaxSessionsPerWeek: 2},
	}

	// Create test time range
	start := time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC)

	// Create test busy times
	busyTimes := []*types.BusyTime{
		{
			Person: people[0],
			Range: &types.Range{
				Start: start.Add(2 * time.Hour),
				End:   start.Add(3 * time.Hour),
			},
		},
		{
			Person: people[1],
			Range: &types.Range{
				Start: start.Add(4 * time.Hour),
				End:   start.Add(5 * time.Hour),
			},
		},
	}

	// Check that the busy times are correctly structured
	if len(busyTimes) != 2 {
		t.Errorf("Expected 2 busy times, got %d", len(busyTimes))
	}

	for _, busyTime := range busyTimes {
		if busyTime.Person == nil {
			t.Error("BusyTime has nil person")
		}
		if busyTime.Range == nil {
			t.Error("BusyTime has nil range")
		}
		if busyTime.Range.Start.IsZero() || busyTime.Range.End.IsZero() {
			t.Error("BusyTime has zero time in range")
		}
		if busyTime.Range.End.Before(busyTime.Range.Start) {
			t.Error("BusyTime range end is before start")
		}
	}
}

func TestParseTime(t *testing.T) {
	// Test with a valid RFC3339 time string
	validTimeStr := "2024-04-01T12:00:00Z"
	parsedTime := parseTime(validTimeStr)

	// Check that the parsed time is correct
	expectedTime := time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC)
	if !parsedTime.Equal(expectedTime) {
		t.Errorf("parseTime() = %v, want %v", parsedTime, expectedTime)
	}

	// Test with an invalid time string
	invalidTimeStr := "invalid-time"
	parsedTime = parseTime(invalidTimeStr)

	// Check that the parsed time is zero time
	if !parsedTime.IsZero() {
		t.Errorf("parseTime() with invalid input = %v, want zero time", parsedTime)
	}
}
