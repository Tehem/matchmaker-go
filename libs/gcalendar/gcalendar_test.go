package gcalendar

import (
	"matchmaker/libs/testutils"
	"matchmaker/libs/types"
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "UTC time",
			input:    time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC),
			expected: "2024-03-15T14:30:00Z",
		},
		{
			name:     "Local time",
			input:    time.Date(2024, 3, 15, 14, 30, 0, 0, time.Local),
			expected: time.Date(2024, 3, 15, 14, 30, 0, 0, time.Local).Format(time.RFC3339),
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTime(tt.input)
			if result != tt.expected {
				t.Errorf("FormatTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFindAvailableSlots_NoCalendars(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create a mock calendar service
	service, err := testutils.MockCalendarService()
	if err != nil {
		t.Fatalf("Failed to create mock calendar service: %v", err)
	}

	// Create a GCalendar instance
	gcal := &GCalendar{service: service}

	// Test parameters
	start := time.Now().Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	duration := 60 // 1 hour in minutes

	// Call the function
	slots, err := gcal.FindAvailableSlots(start, end, []string{}, duration)
	if err != nil {
		t.Fatalf("FindAvailableSlots() error = %v", err)
	}

	// Check that slots are not nil
	if slots == nil {
		t.Error("FindAvailableSlots() returned nil slots")
	}

	// Check that slots are within the expected time range
	for _, slot := range slots {
		if slot.Start.Before(start) || slot.End.After(end) {
			t.Errorf("FindAvailableSlots() returned slot outside range: %v - %v", slot.Start, slot.End)
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

	// Create a mock calendar service
	service, err := testutils.MockCalendarService()
	if err != nil {
		t.Fatalf("Failed to create mock calendar service: %v", err)
	}

	// Create a GCalendar instance
	gcal := &GCalendar{service: service}

	// Test parameters
	session := &types.ReviewSession{
		Reviewers: &types.Squad{
			People: []*types.Person{
				{Email: "reviewer1@example.com"},
				{Email: "reviewer2@example.com"},
			},
		},
		Range: &types.Range{
			Start: time.Now().Add(1 * time.Hour),
			End:   time.Now().Add(2 * time.Hour),
		},
	}

	// Call the function
	_, err = gcal.CreateSessionEvent(session)
	if err != nil {
		t.Fatalf("CreateSessionEvent() error = %v", err)
	}
}

func TestGetBusyTimesForPeople(t *testing.T) {
	// Create a config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	defer configMock.Restore()

	// Create a mock calendar service
	service, err := testutils.MockCalendarService()
	if err != nil {
		t.Fatalf("Failed to create mock calendar service: %v", err)
	}

	// Create a GCalendar instance
	gcal := &GCalendar{service: service}

	// Test parameters
	start := time.Now().Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	people := []*types.Person{
		{Email: "test@example.com", MaxSessionsPerWeek: 2},
	}
	workRanges := []*types.Range{
		{
			Start: start,
			End:   end,
		},
	}

	// Call the function
	busyTimes := gcal.GetBusyTimesForPeople(people, workRanges)

	// Check that busyTimes is not nil
	if busyTimes == nil {
		t.Error("GetBusyTimesForPeople() returned nil busyTimes")
	}
}

func TestParseTime(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "Valid RFC3339 time",
			input:    "2024-03-15T14:30:00Z",
			expected: time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Invalid time format",
			input:    "invalid-time",
			expected: time.Time{},
			wantErr:  true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTime(tt.input)
			if tt.wantErr && !result.IsZero() {
				t.Errorf("parseTime() = %v, expected zero time for invalid input", result)
			}
			if !tt.wantErr && !result.Equal(tt.expected) {
				t.Errorf("parseTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}
