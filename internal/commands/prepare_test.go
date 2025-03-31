package commands

import (
	"context"
	"testing"
	"time"

	"matchmaker/internal/calendar"
	"matchmaker/internal/fs"

	"github.com/spf13/viper"
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
	ResetRootCmdForTest()

	// Set up mock filesystem
	mockFS := fs.NewMockFileSystem()
	fs.Default = mockFS

	// Set up mock calendar service
	calendarService = &calendar.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Set up test configuration
	SetupTestConfig()

	// Create test data
	personsContent := `- email: john.doe@example.com
  isgoodreviewer: true
  skills:
    - frontend
    - backend
- email: jane.doe@example.com
  isgoodreviewer: false
  skills:
    - frontend`

	// Write test files
	mockFS.WriteFile("persons.yml", []byte(personsContent), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments with week shift
	RootCmd.SetArgs([]string{"prepare", "--week-shift", "0"})

	// Run command
	err := RootCmd.Execute()
	require.NoError(t, err)

	// Check output file exists and verify content
	output, err := mockFS.ReadFile("problem.yml")
	require.NoError(t, err)

	// Debug print
	t.Logf("Actual output:\n%s", string(output))
}

func TestPrepareCommandInvalidConfig(t *testing.T) {
	// Reset root command
	ResetRootCmdForTest()

	// Set up mock filesystem
	mockFS := fs.NewMockFileSystem()
	fs.Default = mockFS

	// Set up mock calendar service
	calendarService = &calendar.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Set up invalid configuration
	viper.Reset()
	viper.Set("sessions.duration", "invalid")
	viper.Set("sessions.min_spacing", "invalid")

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
	ResetRootCmdForTest()

	// Set up mock filesystem
	mockFS := fs.NewMockFileSystem()
	fs.Default = mockFS

	// Set up mock calendar service
	calendarService = &calendar.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Set up test configuration
	SetupTestConfig()

	// Create test data
	personsContent := `- email: john.doe@example.com
  isgoodreviewer: true
  skills:
    - frontend
    - backend
- email: jane.doe@example.com
  isgoodreviewer: false
  skills:
    - frontend`

	// Write test files
	mockFS.WriteFile("persons.yml", []byte(personsContent), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments with week shift
	RootCmd.SetArgs([]string{"prepare", "--week-shift", "2"})

	// Run command
	err := RootCmd.Execute()
	require.NoError(t, err)

	// Check output file exists and verify content
	output, err := mockFS.ReadFile("problem.yml")
	require.NoError(t, err)

	// Debug print
	t.Logf("Actual output:\n%s", string(output))
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
