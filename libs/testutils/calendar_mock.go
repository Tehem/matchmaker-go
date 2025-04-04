package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// MockCalendarService creates a mock calendar service for testing
func MockCalendarService() (*calendar.Service, error) {
	// Create a mock HTTP client that returns empty responses
	mockClient := &http.Client{
		Transport: &mockTransport{},
	}

	// Create a new service with the mock client
	ctx := context.Background()
	service, err := calendar.NewService(ctx, option.WithHTTPClient(mockClient))
	if err != nil {
		return nil, err
	}

	return service, nil
}

// mockTransport is a mock HTTP transport that returns empty responses
type mockTransport struct{}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create a response based on the request URL
	var respBody []byte
	var err error

	switch {
	case req.URL.Path == "/freeBusy":
		// Return an empty free/busy response
		resp := &calendar.FreeBusyResponse{
			Calendars: map[string]calendar.FreeBusyCalendar{},
			Kind:      "calendar#freeBusy",
			TimeMin:   time.Now().Format(time.RFC3339),
			TimeMax:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		}
		respBody, err = json.Marshal(resp)
		if err != nil {
			return nil, err
		}

	case req.URL.Path == "/calendars/primary/events":
		// Return an empty events list
		resp := &calendar.Events{
			Items: []*calendar.Event{},
			Kind:  "calendar#events",
		}
		respBody, err = json.Marshal(resp)
		if err != nil {
			return nil, err
		}

	default:
		// Return an empty response for other requests
		respBody = []byte("{}")
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(respBody)),
		Header:     make(http.Header),
	}, nil
}

// CreateMockOAuthConfig creates a mock OAuth2 configuration for testing
func CreateMockOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080",
		Scopes:       []string{calendar.CalendarScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}
}

// CreateMockToken creates a mock OAuth2 token for testing
func CreateMockToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
}

// CreateMockEvent creates a mock calendar event for testing
func CreateMockEvent(summary string, start, end time.Time, attendees []string) *calendar.Event {
	event := &calendar.Event{
		Summary: summary,
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		Attendees: make([]*calendar.EventAttendee, len(attendees)),
		Reminders: &calendar.EventReminders{
			UseDefault: false,
			Overrides: []*calendar.EventReminder{
				{Method: "email", Minutes: 24 * 60},
				{Method: "popup", Minutes: 10},
			},
		},
	}

	for i, email := range attendees {
		event.Attendees[i] = &calendar.EventAttendee{
			Email: email,
		}
	}

	return event
}
