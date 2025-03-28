package commands

import (
	"context"
	"testing"
	"time"

	"matchmaker/internal/config"
	"matchmaker/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock time.Now for testing
var mockNow = time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC) // Wednesday, March 20, 2024

func init() {
	// Override time.Now for testing
	timeNow = func() time.Time {
		return mockNow
	}
}

func TestPrepareCommand(t *testing.T) {
	// Reset root command
	resetRootCmdForTest()

	// Set up mock filesystem
	fs := testutil.NewMockFileSystem()
	config.SetFileSystem(fs)
	defer config.SetFileSystem(config.DefaultFileSystem{})

	// Set up mock calendar service
	calendarService = &testutil.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Create test data
	fs.WriteFile("configs/config.yml", []byte(`sessions:
  duration: 60m
  min_spacing: 2h
  max_per_person_per_week: 3
  session_prefix: "Review Session"
calendar:
  work_hours:
    start: "09:00"
    end: "17:00"
  timezone: "UTC"
  working_days:
    - "Monday"
    - "Tuesday"
    - "Wednesday"
    - "Thursday"
    - "Friday"`), 0644)

	fs.WriteFile("persons.yml", []byte(`- email: john.doe@example.com
  isgoodreviewer: true
  skills:
    - frontend
    - backend
- email: jane.doe@example.com
  isgoodreviewer: false
  skills:
    - frontend`), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments with week shift
	RootCmd.SetArgs([]string{"prepare", "--week-shift", "0"})

	// Run command
	err := RootCmd.Execute()
	require.NoError(t, err)

	// Check output file exists and verify content
	output, err := fs.ReadFile("problem.yml")
	require.NoError(t, err)
	assert.Contains(t, string(output), "targetWeek")
	assert.Contains(t, string(output), "2024-03-25") // Next Monday
}

func TestPrepareCommandInvalidConfig(t *testing.T) {
	// Reset root command
	resetRootCmdForTest()

	// Set up mock filesystem
	fs := testutil.NewMockFileSystem()
	config.SetFileSystem(fs)
	defer config.SetFileSystem(config.DefaultFileSystem{})

	// Set up mock calendar service
	calendarService = &testutil.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Create invalid config
	fs.WriteFile("configs/config.yml", []byte(`sessions:
  duration: invalid
  min_spacing: invalid`), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments with week shift
	RootCmd.SetArgs([]string{"prepare", "--week-shift", "0"})

	// Run command and expect error
	err := RootCmd.Execute()
	assert.Error(t, err)
}

func TestPrepareCommandWithWeekShift(t *testing.T) {
	// Reset root command
	resetRootCmdForTest()

	// Set up mock filesystem
	fs := testutil.NewMockFileSystem()
	config.SetFileSystem(fs)
	defer config.SetFileSystem(config.DefaultFileSystem{})

	// Set up mock calendar service
	calendarService = &testutil.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Create test data
	fs.WriteFile("configs/config.yml", []byte(`sessions:
  duration: 60m
  min_spacing: 2h
  max_per_person_per_week: 3
  session_prefix: "Review Session"
calendar:
  work_hours:
    start: "09:00"
    end: "17:00"
  timezone: "UTC"
  working_days:
    - "Monday"
    - "Tuesday"
    - "Wednesday"
    - "Thursday"
    - "Friday"`), 0644)

	fs.WriteFile("persons.yml", []byte(`- email: john.doe@example.com
  isgoodreviewer: true
  skills:
    - frontend
    - backend`), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments with week shift
	RootCmd.SetArgs([]string{"prepare", "--week-shift", "1"})

	// Run command
	err := RootCmd.Execute()
	require.NoError(t, err)

	// Check output file exists and verify content
	output, err := fs.ReadFile("problem.yml")
	require.NoError(t, err)
	assert.Contains(t, string(output), "targetWeek")
	assert.Contains(t, string(output), "2024-04-01") // Monday after next
}

func TestCalculateTargetWeek(t *testing.T) {
	// Use a fixed date for testing
	now := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC) // Wednesday, March 20, 2024

	// Test with current week (shift = 0)
	nextWeek := calculateTargetWeekFromDate(now, 0)
	assert.Equal(t, "2024-03-25", nextWeek.Format("2006-01-02")) // Next Monday

	// Test with next week (shift = 1)
	nextNextWeek := calculateTargetWeekFromDate(now, 1)
	assert.Equal(t, "2024-04-01", nextNextWeek.Format("2006-01-02")) // Monday after next

	// Test with previous week (shift = -1)
	previousWeek := calculateTargetWeekFromDate(now, -1)
	assert.Equal(t, "2024-03-18", previousWeek.Format("2006-01-02")) // Previous Monday
}
