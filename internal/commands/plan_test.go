package commands

import (
	"context"
	"testing"

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
	calendarService = &testutil.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Create test data
	planningContent := `matches:
    - reviewer1:
        email: john.doe@example.com
        isgoodreviewer: true
        skills:
          - frontend
          - backend
      reviewer2:
        email: jane.doe@example.com
        isgoodreviewer: false
        skills:
          - frontend
      time_slot:
        start: 2024-03-25T10:00:00Z
        end: 2024-03-25T11:00:00Z
      common_skills:
        - frontend`

	configContent := `{
		"sessions": {
			"duration": "60m",
			"min_spacing": "2h",
			"max_per_person_per_week": 3,
			"session_prefix": "Review Session"
		},
		"calendar": {
			"work_hours": {
				"start": "09:00",
				"end": "17:00"
			},
			"timezone": "UTC",
			"working_days": [
				"Monday",
				"Tuesday",
				"Wednesday",
				"Thursday",
				"Friday"
			]
		}
	}`

	// Write test files
	fs.WriteFile("planning.yml", []byte(planningContent), 0644)
	fs.WriteFile("configs/config.json", []byte(configContent), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments
	RootCmd.SetArgs([]string{"plan"})

	// Run command
	err := RootCmd.Execute()
	require.NoError(t, err)
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
	configContent := `{
		"sessions": {
			"duration": "invalid",
			"min_spacing": "invalid"
		}
	}`

	fs.WriteFile("configs/config.json", []byte(configContent), 0644)

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
