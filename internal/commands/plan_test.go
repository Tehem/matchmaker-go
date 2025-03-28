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

func TestPlanCommand(t *testing.T) {
	// Reset root command
	resetRootCmdForTest()

	// Set up mock filesystem
	fs := testutil.NewMockFileSystem()
	config.SetFileSystem(fs)
	defer config.SetFileSystem(config.DefaultFileSystem{})

	// Set up mock calendar service
	mockService := &testutil.MockCalendarService{}
	calendarService = mockService
	defer func() { calendarService = nil }()

	// Create test data
	planningContent := `matches:
- reviewer1:
    email: bob.doe@example.com
    isgoodreviewer: true
    maxsessionsperweek: 3
    skills:
    - frontend
    - backend
    freeslots:
    - start: "2024-03-25T10:00:00Z"
      end: "2024-03-25T11:00:00Z"
  reviewer2:
    email: jane.doe@example.com
    isgoodreviewer: false
    maxsessionsperweek: 3
    skills:
    - frontend
    - backend
    freeslots:
    - start: "2024-03-25T10:00:00Z"
      end: "2024-03-25T11:00:00Z"
  time_slot:
    start: "2024-03-25T10:00:00Z"
    end: "2024-03-25T11:00:00Z"
  common_skills:
  - frontend
  - backend`

	configContent := `sessions:
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
    - "Friday"`

	// Write test files
	fs.WriteFile("planning.yml", []byte(planningContent), 0644)
	fs.WriteFile("configs/config.yml", []byte(configContent), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments
	RootCmd.SetArgs([]string{"plan"})

	// Run command
	err := RootCmd.Execute()
	require.NoError(t, err)

	// Verify that the correct events were created
	require.Len(t, mockService.CreatedEvents, 2, "Expected 2 events to be created")

	// Verify first event
	event1 := mockService.CreatedEvents[0]
	assert.Equal(t, "bob.doe@example.com", event1.Email)
	assert.Equal(t, "Review Sessionbob.doe@example.com & jane.doe@example.com", event1.Event.Summary)
	assert.Equal(t, "2024-03-25T10:00:00Z", event1.Event.Start.Format(time.RFC3339))
	assert.Equal(t, "2024-03-25T11:00:00Z", event1.Event.End.Format(time.RFC3339))
	assert.Equal(t, "Code review session\nCommon skills: [frontend backend]", event1.Event.Description)
	assert.Len(t, event1.Event.Attendees, 2)
	assert.Equal(t, "bob.doe@example.com", event1.Event.Attendees[0])
	assert.Equal(t, "jane.doe@example.com", event1.Event.Attendees[1])

	// Verify second event
	event2 := mockService.CreatedEvents[1]
	assert.Equal(t, "jane.doe@example.com", event2.Email)
	assert.Equal(t, "Review Sessionbob.doe@example.com & jane.doe@example.com", event2.Event.Summary)
	assert.Equal(t, "2024-03-25T10:00:00Z", event2.Event.Start.Format(time.RFC3339))
	assert.Equal(t, "2024-03-25T11:00:00Z", event2.Event.End.Format(time.RFC3339))
	assert.Equal(t, "Code review session\nCommon skills: [frontend backend]", event2.Event.Description)
	assert.Len(t, event2.Event.Attendees, 2)
	assert.Equal(t, "bob.doe@example.com", event2.Event.Attendees[0])
	assert.Equal(t, "jane.doe@example.com", event2.Event.Attendees[1])
}

func TestPlanCommandInvalidConfig(t *testing.T) {
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
	configContent := `sessions:
  duration: invalid
  min_spacing: invalid`

	fs.WriteFile("configs/config.yml", []byte(configContent), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments
	RootCmd.SetArgs([]string{"plan"})

	// Run command and expect error
	err := RootCmd.Execute()
	assert.Error(t, err)
}

func TestPlanCommandMissingPlanningFile(t *testing.T) {
	// Reset root command
	resetRootCmdForTest()

	// Set up mock filesystem
	fs := testutil.NewMockFileSystem()
	config.SetFileSystem(fs)
	defer config.SetFileSystem(config.DefaultFileSystem{})

	// Set up mock calendar service
	calendarService = &testutil.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments
	RootCmd.SetArgs([]string{"plan"})

	// Run command and expect error
	err := RootCmd.Execute()
	assert.Error(t, err)
}

func TestPlanCommandInvalidPlanningFile(t *testing.T) {
	// Reset root command
	resetRootCmdForTest()

	// Set up mock filesystem
	fs := testutil.NewMockFileSystem()
	config.SetFileSystem(fs)
	defer config.SetFileSystem(config.DefaultFileSystem{})

	// Set up mock calendar service
	calendarService = &testutil.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Create invalid planning file
	planningContent := `invalid: yaml: content`
	fs.WriteFile("planning.yml", []byte(planningContent), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments
	RootCmd.SetArgs([]string{"plan"})

	// Run command and expect error
	err := RootCmd.Execute()
	assert.Error(t, err)
}
