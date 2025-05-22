package holidays

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Country represents a country for which we can check holidays
type Country string

const (
	France Country = "FR"
	// Add more countries as needed
)

// Holiday represents a holiday with its name and date
type Holiday struct {
	Name string
	Date time.Time
}

// NagerHoliday represents the holiday data structure from the Nager.Date API
type NagerHoliday struct {
	Date        string   `json:"date"`
	Name        string   `json:"name"`
	LocalName   string   `json:"localName"`
	CountryCode string   `json:"countryCode"`
	Fixed       bool     `json:"fixed"`
	Global      bool     `json:"global"`
	Counties    []string `json:"counties,omitempty"`
	LaunchYear  int      `json:"launchYear,omitempty"`
	Type        string   `json:"type"`
}

// apiURL is the base URL for the Nager.Date API
var apiURL = "https://date.nager.at/api/v3/PublicHolidays/%d/%s"

// HolidayGetter is a function type for getting holidays
type HolidayGetter func(country Country, start, end time.Time) ([]Holiday, error)

// DefaultHolidayGetter is the function used to get holidays
var DefaultHolidayGetter HolidayGetter = getHolidaysForRange

// getHolidaysForRange is the internal implementation of GetHolidaysForRange
func getHolidaysForRange(country Country, start, end time.Time) ([]Holiday, error) {
	year := start.Year()
	if end.Year() != year {
		return nil, fmt.Errorf("start and end dates must be in the same year")
	}

	// Call the Nager.Date API
	url := fmt.Sprintf(apiURL, year, country)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch holidays: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var nagerHolidays []NagerHoliday
	if err := json.NewDecoder(resp.Body).Decode(&nagerHolidays); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	// Convert Nager holidays to our Holiday type and filter by date range
	var holidays []Holiday
	for _, nh := range nagerHolidays {
		date, err := time.Parse("2006-01-02", nh.Date)
		if err != nil {
			continue // Skip invalid dates
		}

		if (date.Equal(start) || date.After(start)) && (date.Equal(end) || date.Before(end)) {
			holidays = append(holidays, Holiday{
				Name: nh.LocalName,
				Date: date,
			})
		}
	}

	return holidays, nil
}

// GetHolidaysForRange returns all holidays for a given country and date range
func GetHolidaysForRange(country Country, start, end time.Time) ([]Holiday, error) {
	return DefaultHolidayGetter(country, start, end)
}

// IsHoliday checks if a given date is a holiday in the specified country
func IsHoliday(country Country, date time.Time) (bool, error) {
	holidays, err := GetHolidaysForRange(country, date, date)
	if err != nil {
		return false, err
	}
	return len(holidays) > 0, nil
}
