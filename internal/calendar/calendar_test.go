package calendar

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

func TestNewCalendarService(t *testing.T) {
	ctx := context.Background()
	service, err := NewCalendarService(ctx, option.WithoutAuthentication())
	require.NoError(t, err)
	assert.NotNil(t, service)
}

func TestGetFreeSlots(t *testing.T) {
	ctx := context.Background()
	service, err := NewCalendarService(ctx, option.WithoutAuthentication())
	require.NoError(t, err)

	startTime := time.Date(2024, 3, 25, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 25, 17, 0, 0, 0, time.UTC)

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
	assert.Len(t, freeSlots, 3)

	// Check the free slots are correct
	expectedSlots := []TimeSlot{
		{
			Start: time.Date(2024, 3, 25, 9, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
		},
		{
			Start: time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
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
	ctx := context.Background()
	service, err := NewCalendarService(ctx, option.WithoutAuthentication())
	require.NoError(t, err)

	startTime := time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 29, 23, 59, 59, 0, time.UTC)

	freeSlots, err := service.GetFreeSlots(ctx, "test@example.com", startTime, endTime, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, freeSlots)

	// Check that slots are within the requested time range
	for _, slot := range freeSlots {
		assert.True(t, slot.Start.After(startTime) || slot.Start.Equal(startTime))
		assert.True(t, slot.End.Before(endTime) || slot.End.Equal(endTime))
	}
}

func TestGetFreeSlotsInvalidTimeRange(t *testing.T) {
	ctx := context.Background()
	service, err := NewCalendarService(ctx, option.WithoutAuthentication())
	require.NoError(t, err)

	startTime := time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 24, 23, 59, 59, 0, time.UTC) // End time before start time

	_, err = service.GetFreeSlots(ctx, "test@example.com", startTime, endTime, nil)
	assert.Error(t, err)
}
