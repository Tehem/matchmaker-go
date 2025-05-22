package util

import (
	"errors"
	"fmt"
	"matchmaker/libs/config"
	"matchmaker/libs/holidays"
	"matchmaker/libs/types"
	"time"
)

const (
	// WorkDaysPerWeek represents the number of working days in a week
	WorkDaysPerWeek = 5
	// HoursPerDay represents the number of hours in a day
	HoursPerDay = 24
	// MinutesPerHour represents the number of minutes in an hour
	MinutesPerHour = 60
)

// WorkHours represents the working hours configuration for a day
type WorkHours struct {
	StartHour   int
	StartMinute int
	EndHour     int
	EndMinute   int
}

// timeNow is a variable that can be overridden in tests
var timeNow = time.Now

// FirstDayOfISOWeek returns the first day (Monday) of the ISO week with the given shift
// weekShift: number of weeks to shift from next week (0 for next week, 1 for the week after, etc.)
func FirstDayOfISOWeek(weekShift int) time.Time {
	date := timeNow()

	// Get to Monday of current week
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
	}

	// Set time to midnight
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	// Add one week to get to next week, then apply the additional shift
	date = date.AddDate(0, 0, 7*(weekShift+1))

	return date
}

// GetWorkRange returns a work range for a specific day
// Returns an error if the time range is invalid (end time before start time)
func GetWorkRange(beginOfWeek time.Time, day int, startHour int, startMinute int, endHour int, endMinute int) (*types.Range, error) {
	if day < 0 || day >= WorkDaysPerWeek {
		return nil, fmt.Errorf("invalid day: %d (must be between 0 and %d)", day, WorkDaysPerWeek-1)
	}

	if startHour < 0 || startHour >= HoursPerDay || endHour < 0 || endHour >= HoursPerDay {
		return nil, fmt.Errorf("invalid hour: must be between 0 and %d", HoursPerDay-1)
	}

	if startMinute < 0 || startMinute >= MinutesPerHour || endMinute < 0 || endMinute >= MinutesPerHour {
		return nil, fmt.Errorf("invalid minute: must be between 0 and %d", MinutesPerHour-1)
	}

	start := time.Date(
		beginOfWeek.Year(),
		beginOfWeek.Month(),
		beginOfWeek.Day()+day,
		startHour,
		startMinute,
		0,
		0,
		beginOfWeek.Location(),
	)
	end := time.Date(
		beginOfWeek.Year(),
		beginOfWeek.Month(),
		beginOfWeek.Day()+day,
		endHour,
		endMinute,
		0,
		0,
		beginOfWeek.Location(),
	)

	if end.Before(start) {
		return nil, errors.New("end time must be after start time")
	}

	return &types.Range{
		Start: start,
		End:   end,
	}, nil
}

// GetWeekWorkRanges returns a channel of work ranges for the week
// Returns an error if the configuration is invalid
func GetWeekWorkRanges(beginOfWeek time.Time) (chan *types.Range, error) {
	ranges := make(chan *types.Range)

	morningHours, err := config.GetWorkHoursConfig("morning")
	if err != nil {
		return nil, fmt.Errorf("invalid morning hours configuration: %w", err)
	}

	afternoonHours, err := config.GetWorkHoursConfig("afternoon")
	if err != nil {
		return nil, fmt.Errorf("invalid afternoon hours configuration: %w", err)
	}

	// Get the country from config
	country := holidays.Country(config.GetCountry())

	// Calculate the end of the week (Friday)
	endOfWeek := beginOfWeek.AddDate(0, 0, WorkDaysPerWeek-1)

	// Fetch all holidays for the week at once
	weekHolidays, err := holidays.GetHolidaysForRange(country, beginOfWeek, endOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch holidays: %w", err)
	}

	// Log holidays found for the week
	if len(weekHolidays) > 0 {
		LogInfo("Holidays found for the week", map[string]interface{}{
			"start":    beginOfWeek.Format("2006-01-02"),
			"end":      endOfWeek.Format("2006-01-02"),
			"country":  country,
			"holidays": weekHolidays,
		})
	} else {
		LogInfo("No holidays found for the week", map[string]interface{}{
			"start":   beginOfWeek.Format("2006-01-02"),
			"end":     endOfWeek.Format("2006-01-02"),
			"country": country,
		})
	}

	// Create a map of holidays for quick lookup
	holidayMap := make(map[string]bool)
	for _, holiday := range weekHolidays {
		// Store the date in YYYY-MM-DD format for comparison
		holidayDate := holiday.Date.Format("2006-01-02")
		holidayMap[holidayDate] = true
		LogInfo("Skipping holiday", map[string]interface{}{
			"date":    holidayDate,
			"holiday": holiday.Name,
			"country": country,
		})
	}

	go func() {
		defer close(ranges)

		for day := 0; day < WorkDaysPerWeek; day++ {
			currentDate := beginOfWeek.AddDate(0, 0, day)
			dateStr := currentDate.Format("2006-01-02")

			// Skip if this day is a holiday
			if holidayMap[dateStr] {
				continue
			}

			// Morning range
			if morningRange, err := GetWorkRange(beginOfWeek, day,
				morningHours.StartHour, morningHours.StartMinute,
				morningHours.EndHour, morningHours.EndMinute); err == nil {
				ranges <- morningRange
			}

			// Afternoon range
			if afternoonRange, err := GetWorkRange(beginOfWeek, day,
				afternoonHours.StartHour, afternoonHours.StartMinute,
				afternoonHours.EndHour, afternoonHours.EndMinute); err == nil {
				ranges <- afternoonRange
			}
		}
	}()

	return ranges, nil
}

// ToSlice converts a channel of ranges to a slice
func ToSlice(c chan *types.Range) []*types.Range {
	s := make([]*types.Range, 0)
	for r := range c {
		s = append(s, r)
	}
	return s
}
