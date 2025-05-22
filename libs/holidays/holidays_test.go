package holidays

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetHolidaysForRange(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response for French holidays in 2024
		holidays := `[
			{
				"date": "2024-01-01",
				"localName": "Jour de l'an",
				"name": "New Year's Day",
				"countryCode": "FR",
				"fixed": true,
				"global": true,
				"type": "Public"
			},
			{
				"date": "2024-04-01",
				"localName": "Lundi de Pâques",
				"name": "Easter Monday",
				"countryCode": "FR",
				"fixed": false,
				"global": true,
				"type": "Public"
			}
		]`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(holidays))
	}))
	defer server.Close()

	// Override the API URL for testing
	originalURL := apiURL
	apiURL = server.URL + "/%d/%s"
	defer func() { apiURL = originalURL }()

	// Test cases
	tests := []struct {
		name     string
		country  Country
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "New Year's Day",
			country:  France,
			start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "Easter Monday",
			country:  France,
			start:    time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "Regular day",
			country:  France,
			start:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			holidays, err := GetHolidaysForRange(tt.country, tt.start, tt.end)
			if err != nil {
				t.Errorf("GetHolidaysForRange() error = %v", err)
				return
			}
			if len(holidays) != tt.expected {
				t.Errorf("GetHolidaysForRange() found %d holidays, want %d", len(holidays), tt.expected)
			}
		})
	}
}

func TestIsHoliday(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response for French holidays in 2024
		holidays := `[
			{
				"date": "2024-01-01",
				"localName": "Jour de l'an",
				"name": "New Year's Day",
				"countryCode": "FR",
				"fixed": true,
				"global": true,
				"type": "Public"
			}
		]`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(holidays))
	}))
	defer server.Close()

	// Override the API URL for testing
	originalURL := apiURL
	apiURL = server.URL + "/%d/%s"
	defer func() { apiURL = originalURL }()

	// Test cases
	tests := []struct {
		name     string
		country  Country
		date     time.Time
		expected bool
	}{
		{
			name:     "New Year's Day",
			country:  France,
			date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "Regular day",
			country:  France,
			date:     time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isHoliday, err := IsHoliday(tt.country, tt.date)
			if err != nil {
				t.Errorf("IsHoliday() error = %v", err)
				return
			}
			if isHoliday != tt.expected {
				t.Errorf("IsHoliday() = %v, want %v", isHoliday, tt.expected)
			}
		})
	}
}
