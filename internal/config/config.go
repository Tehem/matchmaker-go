package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// FileSystem defines the interface for file operations
type FileSystem interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// DefaultFileSystem implements FileSystem using the real filesystem
type DefaultFileSystem struct{}

func (DefaultFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (DefaultFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

var fs FileSystem = DefaultFileSystem{}

// SetFileSystem sets the filesystem to use for file operations
func SetFileSystem(newFS FileSystem) {
	fs = newFS
}

// Config represents the application configuration
type Config struct {
	Sessions SessionConfig  `yaml:"sessions"`
	Calendar CalendarConfig `yaml:"calendar"`
}

// SessionConfig represents the configuration for review sessions
type SessionConfig struct {
	Duration            time.Duration `yaml:"duration"`
	MinSpacing          time.Duration `yaml:"min_spacing"`
	MaxPerPersonPerWeek int           `yaml:"max_per_person_per_week"`
	SessionPrefix       string        `yaml:"session_prefix"`
}

// CalendarConfig represents the calendar configuration
type CalendarConfig struct {
	WorkHours   WorkHoursConfig `yaml:"work_hours"`
	Timezone    string          `yaml:"timezone"`
	WorkingDays []string        `yaml:"working_days"`
}

// WorkHoursConfig represents the working hours configuration
type WorkHoursConfig struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := fs.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set default values
	if cfg.Sessions.Duration == 0 {
		cfg.Sessions.Duration = 60 * time.Minute
	}
	if cfg.Sessions.MinSpacing == 0 {
		cfg.Sessions.MinSpacing = 2 * time.Hour
	}
	if cfg.Sessions.MaxPerPersonPerWeek == 0 {
		cfg.Sessions.MaxPerPersonPerWeek = 3
	}
	if cfg.Sessions.SessionPrefix == "" {
		cfg.Sessions.SessionPrefix = "Review Session"
	}
	if cfg.Calendar.WorkHours.Start == "" {
		cfg.Calendar.WorkHours.Start = "09:00"
	}
	if cfg.Calendar.WorkHours.End == "" {
		cfg.Calendar.WorkHours.End = "17:00"
	}
	if len(cfg.Calendar.WorkingDays) == 0 {
		cfg.Calendar.WorkingDays = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// validate validates the configuration
func (c *Config) validate() error {
	if c.Sessions.Duration <= 0 {
		return fmt.Errorf("invalid session duration: %v", c.Sessions.Duration)
	}
	if c.Sessions.MinSpacing <= 0 {
		return fmt.Errorf("invalid min spacing: %v", c.Sessions.MinSpacing)
	}
	if c.Sessions.MaxPerPersonPerWeek <= 0 {
		return fmt.Errorf("invalid max per person per week: %v", c.Sessions.MaxPerPersonPerWeek)
	}
	if c.Sessions.SessionPrefix == "" {
		return fmt.Errorf("session prefix is required")
	}
	if c.Calendar.Timezone == "" {
		return fmt.Errorf("timezone is required")
	}
	if len(c.Calendar.WorkingDays) == 0 {
		return fmt.Errorf("working days are required")
	}

	// Validate time format
	if _, err := time.Parse("15:04", c.Calendar.WorkHours.Start); err != nil {
		return fmt.Errorf("invalid start time format: %v", c.Calendar.WorkHours.Start)
	}
	if _, err := time.Parse("15:04", c.Calendar.WorkHours.End); err != nil {
		return fmt.Errorf("invalid end time format: %v", c.Calendar.WorkHours.End)
	}

	return nil
}
