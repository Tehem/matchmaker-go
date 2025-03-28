package config

import (
	"testing"
	"time"

	"matchmaker/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	SetFileSystem(fs)
	defer SetFileSystem(DefaultFileSystem{})

	fs.WriteFile("config.yml", []byte(`sessions:
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

	// Test loading
	cfg, err := LoadConfig("config.yml")
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check sessions configuration
	assert.Equal(t, 60*time.Minute, cfg.Sessions.Duration)
	assert.Equal(t, 2*time.Hour, cfg.Sessions.MinSpacing)
	assert.Equal(t, 3, cfg.Sessions.MaxPerPersonPerWeek)
	assert.Equal(t, "Review Session", cfg.Sessions.SessionPrefix)

	// Check calendar configuration
	assert.Equal(t, "09:00", cfg.Calendar.WorkHours.Start)
	assert.Equal(t, "17:00", cfg.Calendar.WorkHours.End)
	assert.Equal(t, "UTC", cfg.Calendar.Timezone)
	assert.Equal(t, []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}, cfg.Calendar.WorkingDays)
}

func TestLoadConfigInvalidFile(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	SetFileSystem(fs)
	defer SetFileSystem(DefaultFileSystem{})

	_, err := LoadConfig("nonexistent.yml")
	assert.Error(t, err)
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	SetFileSystem(fs)
	defer SetFileSystem(DefaultFileSystem{})

	fs.WriteFile("config.yml", []byte(`sessions:
  duration: invalid
  min_spacing: invalid
  max_per_person_per_week: invalid`), 0644)

	// Test loading invalid YAML
	_, err := LoadConfig("config.yml")
	assert.Error(t, err)
}

func TestLoadConfigMissingFields(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	SetFileSystem(fs)
	defer SetFileSystem(DefaultFileSystem{})

	fs.WriteFile("config.yml", []byte(`sessions:
  duration: 60m
calendar:
  timezone: "UTC"`), 0644)

	// Test loading with missing fields
	cfg, err := LoadConfig("config.yml")
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check default values for missing fields
	assert.Equal(t, 60*time.Minute, cfg.Sessions.Duration)
	assert.Equal(t, 2*time.Hour, cfg.Sessions.MinSpacing)
	assert.Equal(t, 3, cfg.Sessions.MaxPerPersonPerWeek)
	assert.Equal(t, "Review Session", cfg.Sessions.SessionPrefix)
	assert.Equal(t, "09:00", cfg.Calendar.WorkHours.Start)
	assert.Equal(t, "17:00", cfg.Calendar.WorkHours.End)
	assert.Equal(t, "UTC", cfg.Calendar.Timezone)
	assert.Equal(t, []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}, cfg.Calendar.WorkingDays)
}

func TestLoadConfigInvalidTimeFormat(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	SetFileSystem(fs)
	defer SetFileSystem(DefaultFileSystem{})

	fs.WriteFile("config.yml", []byte(`sessions:
  duration: 60m
calendar:
  work_hours:
    start: "25:00"
    end: "26:00"
  timezone: "UTC"`), 0644)

	// Test loading with invalid time format
	_, err := LoadConfig("config.yml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid start time format")
}
