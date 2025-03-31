package commands

import (
	"context"
	"testing"

	"matchmaker/internal/calendar"
	"matchmaker/internal/fs"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanCommand(t *testing.T) {
	// Reset root command
	ResetRootCmdForTest()

	// Save the original filesystem and restore it after the test
	originalFS := fs.Default
	defer func() { fs.Default = originalFS }()

	// Set up mock filesystem
	mockFS := fs.NewMockFileSystem()
	fs.Default = mockFS

	// Set up mock calendar service
	calendarService = &calendar.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Set up test configuration
	SetupTestConfig()

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

	// Write test files
	mockFS.WriteFile("planning.yml", []byte(planningContent), 0644)

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
	ResetRootCmdForTest()

	// Save the original filesystem and restore it after the test
	originalFS := fs.Default
	defer func() { fs.Default = originalFS }()

	// Set up mock filesystem
	mockFS := fs.NewMockFileSystem()
	fs.Default = mockFS

	// Set up mock calendar service
	calendarService = &calendar.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Set up invalid configuration
	viper.Reset()
	viper.Set("sessions.session_prefix", "")

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
	ResetRootCmdForTest()

	// Save the original filesystem and restore it after the test
	originalFS := fs.Default
	defer func() { fs.Default = originalFS }()

	// Set up mock filesystem
	mockFS := fs.NewMockFileSystem()
	fs.Default = mockFS

	// Set up mock calendar service
	calendarService = &calendar.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Set up test configuration
	SetupTestConfig()

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
	ResetRootCmdForTest()

	// Save the original filesystem and restore it after the test
	originalFS := fs.Default
	defer func() { fs.Default = originalFS }()

	// Set up mock filesystem
	mockFS := fs.NewMockFileSystem()
	fs.Default = mockFS

	// Set up mock calendar service
	calendarService = &calendar.MockCalendarService{}
	defer func() { calendarService = nil }()

	// Set up test configuration
	SetupTestConfig()

	// Write invalid YAML file
	mockFS.WriteFile("planning.yml", []byte(`invalid: yaml: content`), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments
	RootCmd.SetArgs([]string{"plan"})

	// Run command and expect error
	err := RootCmd.Execute()
	assert.Error(t, err)
}
