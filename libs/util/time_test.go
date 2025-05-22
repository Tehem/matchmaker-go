package util

import (
	"matchmaker/libs/holidays"
	"matchmaker/libs/testutils"
	"matchmaker/libs/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFirstDayOfISOWeek(t *testing.T) {
	// Mock the current time to a known date (Wednesday, March 20, 2024)
	mockTime := time.Date(2024, 3, 20, 15, 30, 0, 0, time.UTC)
	timeNow = func() time.Time { return mockTime }
	defer func() { timeNow = time.Now }() // Restore original timeNow function

	// Test next week (shift 0)
	result := FirstDayOfISOWeek(0)
	expected := time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC) // Monday, March 25, 2024
	assert.Equal(t, expected, result, "FirstDayOfISOWeek(0) should return Monday of next week")

	// Test week after next (shift 1)
	result = FirstDayOfISOWeek(1)
	expected = time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC) // Monday, April 1, 2024
	assert.Equal(t, expected, result, "FirstDayOfISOWeek(1) should return Monday of the week after next")

	// Test two weeks after next (shift 2)
	result = FirstDayOfISOWeek(2)
	expected = time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC) // Monday, April 8, 2024
	assert.Equal(t, expected, result, "FirstDayOfISOWeek(2) should return Monday of two weeks after next")

	// Test three weeks after next (shift 3)
	result = FirstDayOfISOWeek(3)
	expected = time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC) // Monday, April 15, 2024
	assert.Equal(t, expected, result, "FirstDayOfISOWeek(3) should return Monday of three weeks after next")
}

// TestFirstDayOfISOWeekWithDifferentStartDays verifies the function works correctly
// regardless of which day of the week we start from
func TestFirstDayOfISOWeekWithDifferentStartDays(t *testing.T) {
	// Test with different days of the week
	days := []time.Weekday{
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
		time.Sunday,
	}

	for _, day := range days {
		// Set mock time to a specific day of the week
		mockTime := time.Date(2024, 3, 20, 15, 30, 0, 0, time.UTC)
		for mockTime.Weekday() != day {
			mockTime = mockTime.AddDate(0, 0, 1)
		}
		timeNow = func() time.Time { return mockTime }

		t.Run(day.String(), func(t *testing.T) {
			// Test with shift 0 (next week)
			result := FirstDayOfISOWeek(0)
			expected := mockTime.AddDate(0, 0, 7) // Move to next week
			for expected.Weekday() != time.Monday {
				expected = expected.AddDate(0, 0, -1)
			}
			expected = time.Date(expected.Year(), expected.Month(), expected.Day(), 0, 0, 0, 0, expected.Location())

			assert.Equal(t, expected, result, "FirstDayOfISOWeek(0) should return Monday of next week when starting from %s", day)
		})
	}

	// Restore original timeNow function
	timeNow = time.Now
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

func TestGetWeekWorkRanges(t *testing.T) {
	// Create mock holidays
	mockHolidays := testutils.NewMockHolidays()
	originalHolidayGetter := holidays.DefaultHolidayGetter
	holidays.DefaultHolidayGetter = func(country holidays.Country, start, end time.Time) ([]holidays.Holiday, error) {
		return mockHolidays.GetHolidaysForRange(country, start, end)
	}
	defer func() {
		holidays.DefaultHolidayGetter = originalHolidayGetter
	}()

	// Setup config mock
	configMock := testutils.NewConfigMock()
	configMock.SetupWorkHours()
	configMock.SetMorningHours(10, 0, 12, 0)
	configMock.SetAfternoonHours(14, 0, 18, 0)
	defer configMock.Restore()

	// Mock the current time to a known date (Wednesday, March 20, 2024)
	mockTime := time.Date(2024, 3, 20, 15, 30, 0, 0, time.UTC)
	timeNow = func() time.Time { return mockTime }
	defer func() { timeNow = time.Now }() // Restore original timeNow function

	// Test case 1: Week with no holidays
	beginOfWeek := FirstDayOfISOWeek(0) // Monday, March 25, 2024
	rangesChan, err := GetWeekWorkRanges(beginOfWeek)
	assert.NoError(t, err)
	ranges := ToSlice(rangesChan)

	// Should have 10 ranges (5 days × 2 ranges per day)
	assert.Equal(t, 10, len(ranges))

	// Verify first and last ranges
	assert.Equal(t, time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC), ranges[0].Start) // Monday morning
	assert.Equal(t, time.Date(2024, 3, 25, 12, 0, 0, 0, time.UTC), ranges[0].End)
	assert.Equal(t, time.Date(2024, 3, 29, 14, 0, 0, 0, time.UTC), ranges[9].Start) // Friday afternoon
	assert.Equal(t, time.Date(2024, 3, 29, 18, 0, 0, 0, time.UTC), ranges[9].End)

	// Test case 2: Week with a holiday (Easter Monday, April 1, 2024)
	beginOfWeek = FirstDayOfISOWeek(1) // Monday, April 1, 2024
	rangesChan, err = GetWeekWorkRanges(beginOfWeek)
	assert.NoError(t, err)
	ranges = ToSlice(rangesChan)

	// Should have 8 ranges (4 days × 2 ranges per day, excluding Easter Monday)
	assert.Equal(t, 8, len(ranges))

	// Verify ranges don't include Easter Monday
	for _, r := range ranges {
		easterMonday := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
		assert.False(t, r.Start.Month() == easterMonday.Month() && r.Start.Day() == easterMonday.Day(),
			"Ranges should not include Easter Monday")
	}

	// Test case 3: Week with multiple holidays (May 1, 2024 - Labor Day)
	beginOfWeek = time.Date(2024, 4, 29, 0, 0, 0, 0, time.UTC) // Monday, April 29, 2024
	rangesChan, err = GetWeekWorkRanges(beginOfWeek)
	assert.NoError(t, err)
	ranges = ToSlice(rangesChan)

	// Should have 8 ranges (4 days × 2 ranges per day, excluding May 1)
	assert.Equal(t, 8, len(ranges))

	// Verify ranges don't include May 1
	for _, r := range ranges {
		assert.NotEqual(t, time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC).Day(), r.Start.Day(),
			"Ranges should not include Labor Day")
	}

	// Test case 4: Week with all holidays (Christmas week)
	beginOfWeek = time.Date(2024, 12, 23, 0, 0, 0, 0, time.UTC) // Monday, December 23, 2024
	rangesChan, err = GetWeekWorkRanges(beginOfWeek)
	assert.NoError(t, err)
	ranges = ToSlice(rangesChan)

	// Should have 6 ranges (3 days × 2 ranges per day, excluding Christmas and Boxing Day)
	assert.Equal(t, 6, len(ranges))

	// Verify ranges don't include December 25 or 26
	for _, r := range ranges {
		day := r.Start.Day()
		assert.NotEqual(t, 25, day, "Ranges should not include Christmas Day")
		assert.NotEqual(t, 26, day, "Ranges should not include Boxing Day")
	}

	// Test case 5: Week with custom holiday
	mockHolidays.ClearHolidays()
	mockHolidays.AddHoliday("Custom Holiday", time.Date(2024, 3, 27, 0, 0, 0, 0, time.UTC)) // Wednesday
	beginOfWeek = FirstDayOfISOWeek(0)                                                      // Monday, March 25, 2024
	rangesChan, err = GetWeekWorkRanges(beginOfWeek)
	assert.NoError(t, err)
	ranges = ToSlice(rangesChan)

	// Should have 8 ranges (4 days × 2 ranges per day, excluding Wednesday)
	assert.Equal(t, 8, len(ranges))

	// Verify ranges don't include Wednesday
	for _, r := range ranges {
		assert.NotEqual(t, time.Date(2024, 3, 27, 0, 0, 0, 0, time.UTC).Day(), r.Start.Day(),
			"Ranges should not include Custom Holiday")
	}
}
