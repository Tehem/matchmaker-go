package util

import (
	"matchmaker/libs/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFirstDayOfISOWeek(t *testing.T) {
	// Set a fixed date for testing
	fixedDate := time.Date(2024, time.April, 3, 15, 30, 0, 0, time.UTC)
	timeNow = func() time.Time {
		return fixedDate
	}
	defer func() { timeNow = time.Now }()

	// Test current week (weekShift=0)
	result := FirstDayOfISOWeek(0)
	assert.Equal(t, time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC), result)

	// Test next week (weekShift=1)
	result = FirstDayOfISOWeek(1)
	assert.Equal(t, time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC), result)

	// Test previous week (weekShift=-1)
	result = FirstDayOfISOWeek(-1)
	assert.Equal(t, time.Date(2024, time.March, 25, 0, 0, 0, 0, time.UTC), result)
}

func TestGetWorkRange(t *testing.T) {
	// Create a base time (Monday)
	baseTime := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC) // Monday, Jan 2, 2023

	// Test valid work range
	range1, err := GetWorkRange(baseTime, 0, 9, 0, 17, 0)
	assert.NoError(t, err)
	assert.NotNil(t, range1)
	assert.Equal(t, time.Date(2023, 1, 2, 9, 0, 0, 0, time.UTC), range1.Start)
	assert.Equal(t, time.Date(2023, 1, 2, 17, 0, 0, 0, time.UTC), range1.End)

	// Test invalid day
	_, err = GetWorkRange(baseTime, -1, 9, 0, 17, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid day")

	_, err = GetWorkRange(baseTime, WorkDaysPerWeek, 9, 0, 17, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid day")

	// Test invalid hour
	_, err = GetWorkRange(baseTime, 0, -1, 0, 17, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hour")

	_, err = GetWorkRange(baseTime, 0, HoursPerDay, 0, 17, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hour")

	// Test invalid minute
	_, err = GetWorkRange(baseTime, 0, 9, -1, 17, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid minute")

	_, err = GetWorkRange(baseTime, 0, 9, MinutesPerHour, 17, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid minute")

	// Test end time before start time
	_, err = GetWorkRange(baseTime, 0, 17, 0, 9, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end time must be after start time")

	// Test different days of the week
	for day := 0; day < WorkDaysPerWeek; day++ {
		range2, err := GetWorkRange(baseTime, day, 9, 0, 17, 0)
		assert.NoError(t, err)
		assert.NotNil(t, range2)
		expectedStart := time.Date(2023, 1, 2+day, 9, 0, 0, 0, time.UTC)
		expectedEnd := time.Date(2023, 1, 2+day, 17, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedStart, range2.Start)
		assert.Equal(t, expectedEnd, range2.End)
	}
}

func TestToSlice(t *testing.T) {
	// Create a channel of ranges
	c := make(chan *types.Range)

	// Create some test ranges
	now := time.Now()
	ranges := []*types.Range{
		{
			Start: now,
			End:   now.Add(time.Hour),
		},
		{
			Start: now.Add(time.Hour * 2),
			End:   now.Add(time.Hour * 3),
		},
		{
			Start: now.Add(time.Hour * 4),
			End:   now.Add(time.Hour * 5),
		},
	}

	// Send ranges to channel in a goroutine
	go func() {
		for _, r := range ranges {
			c <- r
		}
		close(c)
	}()

	// Convert channel to slice
	result := ToSlice(c)

	// Verify result
	assert.Equal(t, len(ranges), len(result))
	for i, r := range ranges {
		assert.Equal(t, r.Start, result[i].Start)
		assert.Equal(t, r.End, result[i].End)
	}
}
