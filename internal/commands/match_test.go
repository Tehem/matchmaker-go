package commands

import (
	"context"
	"testing"

	"matchmaker/internal/config"
	"matchmaker/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchCommand(t *testing.T) {
	// Reset root command
	resetRootCmdForTest()

	// Set up mock filesystem
	fs := testutil.NewMockFileSystem()
	config.SetFileSystem(fs)
	defer config.SetFileSystem(config.DefaultFileSystem{})

	// Create test data with a person who has max sessions per week set to 0
	problemContent := `target_week: "2024-03-25T00:00:00Z"
people:
  - email: john.doe@example.com
    isgoodreviewer: true
    skills:
      - frontend
      - backend
    maxsessionsperweek: 0
    freeslots:
      - start: "2024-03-25T10:00:00Z"
        end: "2024-03-25T11:00:00Z"
  - email: jane.doe@example.com
    isgoodreviewer: false
    skills:
      - frontend
      - backend
    maxsessionsperweek: 3
    freeslots:
      - start: "2024-03-25T10:00:00Z"
        end: "2024-03-25T11:00:00Z"
  - email: bob.doe@example.com
    isgoodreviewer: true
    skills:
      - frontend
      - backend
    maxsessionsperweek: 3
    freeslots:
      - start: "2024-03-25T10:00:00Z"
        end: "2024-03-25T11:00:00Z"`

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
	fs.WriteFile("problem.yml", []byte(problemContent), 0644)
	fs.WriteFile("configs/config.yml", []byte(configContent), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments
	RootCmd.SetArgs([]string{"match"})

	// Run command
	err := RootCmd.Execute()
	require.NoError(t, err)

	// Check output file exists and verify content
	output, err := fs.ReadFile("planning.yml")
	require.NoError(t, err)

	// Debug print
	t.Logf("Actual output:\n%s", string(output))

	// Verify the output format and content
	expectedContent := `matches:
    - reviewer1:
        email: bob.doe@example.com
        isgoodreviewer: true
        maxsessionsperweek: 3
        skills:
            - frontend
            - backend
        freeslots:
            - start: 2024-03-25T10:00:00Z
              end: 2024-03-25T11:00:00Z
      reviewer2:
        email: jane.doe@example.com
        isgoodreviewer: false
        maxsessionsperweek: 3
        skills:
            - frontend
            - backend
        freeslots:
            - start: 2024-03-25T10:00:00Z
              end: 2024-03-25T11:00:00Z
      time_slot:
        start: 2024-03-25T10:00:00Z
        end: 2024-03-25T11:00:00Z
      common_skills:
        - frontend
        - backend
`
	assert.Equal(t, expectedContent, string(output))
}

func TestMatchCommandInvalidConfig(t *testing.T) {
	// Reset root command
	resetRootCmdForTest()

	// Set up mock filesystem
	fs := testutil.NewMockFileSystem()
	config.SetFileSystem(fs)
	defer config.SetFileSystem(config.DefaultFileSystem{})

	// Create invalid config
	configContent := `sessions:
  duration: invalid
  min_spacing: invalid`

	fs.WriteFile("configs/config.yml", []byte(configContent), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments
	RootCmd.SetArgs([]string{"match"})

	// Run command and expect error
	err := RootCmd.Execute()
	assert.Error(t, err)
}

func TestMatchCommandMissingProblemFile(t *testing.T) {
	// Reset root command
	resetRootCmdForTest()

	// Set up mock filesystem
	fs := testutil.NewMockFileSystem()
	config.SetFileSystem(fs)
	defer config.SetFileSystem(config.DefaultFileSystem{})

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments
	RootCmd.SetArgs([]string{"match"})

	// Run command and expect error
	err := RootCmd.Execute()
	assert.Error(t, err)
}

func TestMatchCommandInvalidProblemFile(t *testing.T) {
	// Reset root command
	resetRootCmdForTest()

	// Set up mock filesystem
	fs := testutil.NewMockFileSystem()
	config.SetFileSystem(fs)
	defer config.SetFileSystem(config.DefaultFileSystem{})

	// Create invalid problem file
	problemContent := `invalid: yaml: content`
	fs.WriteFile("problem.yml", []byte(problemContent), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments
	RootCmd.SetArgs([]string{"match"})

	// Run command and expect error
	err := RootCmd.Execute()
	assert.Error(t, err)
}

func TestMatchCommandInsufficientPeople(t *testing.T) {
	// Reset root command
	resetRootCmdForTest()

	// Set up mock filesystem
	fs := testutil.NewMockFileSystem()
	config.SetFileSystem(fs)
	defer config.SetFileSystem(config.DefaultFileSystem{})

	// Create test data with only one person
	problemContent := `target_week: "2024-03-25T00:00:00Z"
people:
  - email: john.doe@example.com
    isgoodreviewer: true
    skills:
      - frontend
      - backend
    freeslots:
      - start: "2024-03-25T10:00:00Z"
        end: "2024-03-25T11:00:00Z"`

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
	fs.WriteFile("problem.yml", []byte(problemContent), 0644)
	fs.WriteFile("configs/config.yml", []byte(configContent), 0644)

	// Set up command with context
	ctx := context.Background()
	RootCmd.SetContext(ctx)

	// Set up command arguments
	RootCmd.SetArgs([]string{"match"})

	// Run command and expect error
	err := RootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "need at least 2 people to create matches")
}
