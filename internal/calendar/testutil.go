package calendar

import (
	"context"
	"matchmaker/internal/fs"
	"os"
	"time"
)

// MockFileSystem is a mock implementation of the filesystem interface for testing
type MockFileSystem struct {
	files map[string][]byte
}

// NewMockFileSystem creates a new mock filesystem
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
	}
}

// ReadFile reads a file from the mock filesystem
func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if data, ok := m.files[path]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

// WriteFile writes a file to the mock filesystem
func (m *MockFileSystem) WriteFile(path string, data []byte, perm int) error {
	m.files[path] = data
	return nil
}

// Default is the default filesystem implementation
var Default fs.FileSystem = fs.NewMockFileSystem()

// MockCalendarService implements a mock calendar service for testing
type MockCalendarService struct {
	CreatedEvents []struct {
		Email string
		Event *Event
	}
}

func (m *MockCalendarService) GetFreeSlots(ctx context.Context, email string, startTime, endTime time.Time, events []*Event) ([]TimeSlot, error) {
	return []TimeSlot{
		{
			Start: time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 11, 0, 0, 0, time.UTC),
		},
	}, nil
}

func (m *MockCalendarService) CreateEvent(ctx context.Context, email string, event *Event) error {
	m.CreatedEvents = append(m.CreatedEvents, struct {
		Email string
		Event *Event
	}{
		Email: email,
		Event: event,
	})
	return nil
}

func (m *MockCalendarService) GetBusySlots(ctx context.Context, email string, startTime, endTime time.Time) ([]TimeSlot, error) {
	// Return some mock busy slots
	return []TimeSlot{
		{
			Start: time.Date(2024, 3, 25, 9, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 10, 0, 0, 0, time.UTC),
		},
		{
			Start: time.Date(2024, 3, 25, 14, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 25, 15, 0, 0, 0, time.UTC),
		},
	}, nil
}

// Ensure MockCalendarService implements CalendarService
var _ CalendarService = &MockCalendarService{}
