package calendar

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

// setupTestWorkingHours configures the working hours for testing
func setupTestWorkingHours() {
	viper.Set("workingHours.timezone", "UTC")
	viper.Set("workingHours.morning.start", "09:00")
	viper.Set("workingHours.morning.end", "12:00")
	viper.Set("workingHours.afternoon.start", "13:00")
	viper.Set("workingHours.afternoon.end", "17:00")
}

func TestGetFreeSlots(t *testing.T) {
	setupTestWorkingHours()
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

func TestGetFreeSlotsNoEvents(t *testing.T) {
	setupTestWorkingHours()
	ctx := context.Background()
	service, err := NewCalendarService(ctx, option.WithoutAuthentication())
	require.NoError(t, err)

	startTime := time.Date(2024, 3, 25, 8, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 25, 18, 0, 0, 0, time.UTC)

	freeSlots, err := service.GetFreeSlots(ctx, "test@example.com", startTime, endTime, nil)
	require.NoError(t, err)
	assert.Len(t, freeSlots, 2)

	// Check that we have one slot for morning and one for afternoon
	expectedSlots := []TimeSlot{
		{
			Start: time.Date(2024, 3, 25, 9, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 12, 0, 0, 0, time.UTC),
		},
		{
			Start: time.Date(2024, 3, 25, 13, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 17, 0, 0, 0, time.UTC),
		},
	}

	for i, slot := range freeSlots {
		assert.Equal(t, expectedSlots[i].Start, slot.Start)
		assert.Equal(t, expectedSlots[i].End, slot.End)
	}
}

func TestGetFreeSlotsInvalidTimeRange(t *testing.T) {
	setupTestWorkingHours()
	ctx := context.Background()
	service, err := NewCalendarService(ctx, option.WithoutAuthentication())
	require.NoError(t, err)

	startTime := time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 24, 23, 59, 59, 0, time.UTC) // End time before start time

	_, err = service.GetFreeSlots(ctx, "test@example.com", startTime, endTime, nil)
	assert.Error(t, err)
}
