package calendar

import (
	"context"
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"testing"
	"time"

	"matchmaker/internal/fs"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// setupGoogleCalendarTest configures the working hours for testing
func setupGoogleCalendarTest() {
	viper.Set("workingHours.timezone", "UTC")
	viper.Set("workingHours.morning.start", "09:00")
	viper.Set("workingHours.morning.end", "12:00")
	viper.Set("workingHours.afternoon.start", "13:00")
	viper.Set("workingHours.afternoon.end", "17:00")
}

func TestNewCalendarService(t *testing.T) {
	ctx := context.Background()
	service, err := NewCalendarService(ctx, option.WithoutAuthentication())
	require.NoError(t, err)
	assert.NotNil(t, service)
}

func TestService_GetFreeSlots(t *testing.T) {
	setupGoogleCalendarTest()
	ctx := context.Background()
	service, err := NewCalendarService(ctx, option.WithoutAuthentication())
	require.NoError(t, err)

	// Test date: March 25, 2024
	startTime := time.Date(2024, 3, 25, 8, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 25, 18, 0, 0, 0, time.UTC)

	// Create test events
	events := []*Event{
		{
			Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
		},
		{
			Start: time.Date(2024, 3, 25, 14, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 15, 0, 0, 0, time.UTC),
		},
	}

	freeSlots, err := service.GetFreeSlots(ctx, "test@example.com", startTime, endTime, events)
	require.NoError(t, err)
	assert.Len(t, freeSlots, 4)

	// Check the free slots are correct
	expectedSlots := []TimeSlot{
		{
			Start: time.Date(2024, 3, 25, 9, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
		},
		{
			Start: time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 12, 0, 0, 0, time.UTC),
		},
		{
			Start: time.Date(2024, 3, 25, 13, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 14, 0, 0, 0, time.UTC),
		},
		{
			Start: time.Date(2024, 3, 25, 15, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 17, 0, 0, 0, time.UTC),
		},
	}

	for i, slot := range freeSlots {
		assert.Equal(t, expectedSlots[i].Start, slot.Start)
		assert.Equal(t, expectedSlots[i].End, slot.End)
	}
}

func TestTransformEventWithOrganizer(t *testing.T) {
	event := &Event{
		Summary:        "Test Event",
		Start:          time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
		End:            time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
		Description:    "Test Description",
		Attendees:      []string{"attendee1@example.com", "attendee2@example.com"},
		OrganizerEmail: "organizer@example.com",
	}

	expected := &calendar.Event{
		Summary:     "Test Event",
		Start:       &calendar.EventDateTime{DateTime: "2024-03-25T10:00:00Z"},
		End:         &calendar.EventDateTime{DateTime: "2024-03-25T11:00:00Z"},
		Description: "Test Description",
		Organizer:   &calendar.EventOrganizer{Email: "organizer@example.com"},
		Attendees: []*calendar.EventAttendee{
			{Email: "attendee1@example.com"},
			{Email: "attendee2@example.com"},
			{Email: "organizer@example.com", Optional: true},
		},
	}

	service := &Service{}
	got, err := service.transformEvent(event)
	require.NoError(t, err)

	assert.Equal(t, expected.Summary, got.Summary)
	assert.Equal(t, expected.Description, got.Description)
	assert.Equal(t, expected.Start.DateTime, got.Start.DateTime)
	assert.Equal(t, expected.End.DateTime, got.End.DateTime)
	assert.Equal(t, expected.Organizer, got.Organizer)
	assert.Len(t, got.Attendees, len(expected.Attendees))

	for i, attendee := range expected.Attendees {
		assert.Equal(t, attendee.Email, got.Attendees[i].Email)
		assert.Equal(t, attendee.Optional, got.Attendees[i].Optional)
	}
}

func TestTransformEventWithoutOrganizer(t *testing.T) {
	event := &Event{
		Summary:     "Test Event",
		Start:       time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
		End:         time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
		Description: "Test Description",
		Attendees:   []string{"attendee1@example.com", "attendee2@example.com"},
	}

	expected := &calendar.Event{
		Summary:     "Test Event",
		Start:       &calendar.EventDateTime{DateTime: "2024-03-25T10:00:00Z"},
		End:         &calendar.EventDateTime{DateTime: "2024-03-25T11:00:00Z"},
		Description: "Test Description",
		Attendees: []*calendar.EventAttendee{
			{Email: "attendee1@example.com"},
			{Email: "attendee2@example.com"},
		},
	}

	service := &Service{}
	got, err := service.transformEvent(event)
	require.NoError(t, err)

	assert.Equal(t, expected.Summary, got.Summary)
	assert.Equal(t, expected.Description, got.Description)
	assert.Equal(t, expected.Start.DateTime, got.Start.DateTime)
	assert.Equal(t, expected.End.DateTime, got.End.DateTime)
	assert.Nil(t, got.Organizer)
	assert.Len(t, got.Attendees, len(expected.Attendees))

	for i, attendee := range expected.Attendees {
		assert.Equal(t, attendee.Email, got.Attendees[i].Email)
		assert.Equal(t, attendee.Optional, got.Attendees[i].Optional)
	}
}

func TestTransformEventWithOrganizerAsAttendee(t *testing.T) {
	event := &Event{
		Summary:        "Test Event",
		Start:          time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
		End:            time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
		Description:    "Test Description",
		Attendees:      []string{"attendee1@example.com", "organizer@example.com"},
		OrganizerEmail: "organizer@example.com",
	}

	expected := &calendar.Event{
		Summary:     "Test Event",
		Start:       &calendar.EventDateTime{DateTime: "2024-03-25T10:00:00Z"},
		End:         &calendar.EventDateTime{DateTime: "2024-03-25T11:00:00Z"},
		Description: "Test Description",
		Organizer:   &calendar.EventOrganizer{Email: "organizer@example.com"},
		Attendees: []*calendar.EventAttendee{
			{Email: "attendee1@example.com"},
			{Email: "organizer@example.com"},
		},
	}

	service := &Service{}
	got, err := service.transformEvent(event)
	require.NoError(t, err)

	assert.Equal(t, expected.Summary, got.Summary)
	assert.Equal(t, expected.Description, got.Description)
	assert.Equal(t, expected.Start.DateTime, got.Start.DateTime)
	assert.Equal(t, expected.End.DateTime, got.End.DateTime)
	assert.Equal(t, expected.Organizer, got.Organizer)
	assert.Len(t, got.Attendees, len(expected.Attendees))

	for i, attendee := range expected.Attendees {
		assert.Equal(t, attendee.Email, got.Attendees[i].Email)
		assert.Equal(t, attendee.Optional, got.Attendees[i].Optional)
	}
}

func TestNewCalendarServiceFromToken(t *testing.T) {
	// Save the original filesystem and restore it after the test
	originalFS := Default
	defer func() { Default = originalFS }()

	// Set up mock filesystem
	mockFS := fs.NewMockFileSystem()
	Default = mockFS

	// Create test token
	token := &oauth2.Token{
		AccessToken:  "test_access_token",
		TokenType:    "Bearer",
		RefreshToken: "test_refresh_token",
	}

	// Get user's home directory
	usr, err := user.Current()
	require.NoError(t, err)

	// Create credentials directory
	credentialsDir := filepath.Join(usr.HomeDir, ".credentials")
	err = os.MkdirAll(credentialsDir, 0700)
	require.NoError(t, err)

	// Write token to file
	tokenFile := filepath.Join(credentialsDir, "calendar-api.json")
	tokenData, err := json.Marshal(token)
	require.NoError(t, err)
	err = mockFS.WriteFile(tokenFile, tokenData, 0600)
	require.NoError(t, err)

	// Create context
	ctx := context.Background()

	// Test creating service
	service, err := NewCalendarServiceFromToken(ctx)
	require.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewCalendarServiceFromTokenMissingFile(t *testing.T) {
	// Save the original filesystem and restore it after the test
	originalFS := Default
	defer func() { Default = originalFS }()

	// Set up mock filesystem
	mockFS := fs.NewMockFileSystem()
	Default = mockFS

	// Create context
	ctx := context.Background()

	// Test creating service with missing token file
	service, err := newCalendarServiceFromTokenWithHome(ctx, "/mock/home", mockFS)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to read token file")
}

func TestNewCalendarServiceFromTokenInvalidToken(t *testing.T) {
	// Save the original filesystem and restore it after the test
	originalFS := Default
	defer func() { Default = originalFS }()

	// Set up mock filesystem
	mockFS := fs.NewMockFileSystem()
	Default = mockFS

	// Create credentials directory in mock filesystem
	credentialsDir := "/mock/home/.credentials"
	err := mockFS.MkdirAll(credentialsDir, 0700)
	require.NoError(t, err)

	// Write invalid token to file
	tokenFile := credentialsDir + "/calendar-api.json"
	err = mockFS.WriteFile(tokenFile, []byte("invalid json"), 0600)
	require.NoError(t, err)

	// Verify the file exists in the mock filesystem
	data, err := mockFS.ReadFile(tokenFile)
	require.NoError(t, err)
	assert.Equal(t, []byte("invalid json"), data)

	// Create context
	ctx := context.Background()

	// Test creating service with invalid token
	service, err := newCalendarServiceFromTokenWithHome(ctx, "/mock/home", mockFS)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to decode token")
}
