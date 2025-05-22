package testutils

import (
	"matchmaker/libs/holidays"
	"time"
)

// MockHolidays is a mock implementation of the holidays package
type MockHolidays struct {
	holidays []holidays.Holiday
}

// NewMockHolidays creates a new mock holidays instance
func NewMockHolidays() *MockHolidays {
	return &MockHolidays{
		holidays: []holidays.Holiday{
			// Easter Monday 2024
			{
				Name: "Easter Monday",
				Date: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			// Labor Day 2024
			{
				Name: "Labor Day",
				Date: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			// Christmas 2024
			{
				Name: "Christmas Day",
				Date: time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			},
			// Boxing Day 2024
			{
				Name: "Boxing Day",
				Date: time.Date(2024, 12, 26, 0, 0, 0, 0, time.UTC),
			},
		},
	}
}

// GetHolidaysForRange returns holidays within the specified date range
func (m *MockHolidays) GetHolidaysForRange(country holidays.Country, start, end time.Time) ([]holidays.Holiday, error) {
	var result []holidays.Holiday
	for _, h := range m.holidays {
		if !h.Date.Before(start) && !h.Date.After(end) {
			result = append(result, h)
		}
	}
	return result, nil
}

// IsHoliday checks if a given date is a holiday
func (m *MockHolidays) IsHoliday(country holidays.Country, date time.Time) (bool, error) {
	holidays, err := m.GetHolidaysForRange(country, date, date)
	if err != nil {
		return false, err
	}
	return len(holidays) > 0, nil
}

// AddHoliday adds a new holiday to the mock
func (m *MockHolidays) AddHoliday(name string, date time.Time) {
	m.holidays = append(m.holidays, holidays.Holiday{
		Name: name,
		Date: date,
	})
}

// ClearHolidays removes all holidays from the mock
func (m *MockHolidays) ClearHolidays() {
	m.holidays = nil
}
