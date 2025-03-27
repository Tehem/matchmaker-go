package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Sessions     SessionConfig      `mapstructure:"sessions"`
	WorkingHours WorkingHoursConfig `mapstructure:"workingHours"`
}

// SessionConfig represents the configuration for review sessions
type SessionConfig struct {
	DurationMinutes     int    `mapstructure:"sessionDurationMinutes"`
	MinSpacingHours     int    `mapstructure:"minSessionSpacingHours"`
	MaxPerPersonPerWeek int    `mapstructure:"maxPerPersonPerWeek"`
	SessionPrefix       string `mapstructure:"sessionPrefix"`
}

// WorkingHoursConfig represents the working hours configuration
type WorkingHoursConfig struct {
	Timezone  string         `mapstructure:"timezone"`
	Morning   TimeSlotConfig `mapstructure:"morning"`
	Afternoon TimeSlotConfig `mapstructure:"afternoon"`
}

// TimeSlotConfig represents a time slot configuration
type TimeSlotConfig struct {
	Start TimeConfig `mapstructure:"start"`
	End   TimeConfig `mapstructure:"end"`
}

// TimeConfig represents a time configuration
type TimeConfig struct {
	Hour   int `mapstructure:"hour"`
	Minute int `mapstructure:"minute"`
}

// LoadConfig loads and validates the configuration
func LoadConfig() (*Config, error) {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validate performs validation on the configuration
func (c *Config) validate() error {
	if err := c.Sessions.validate(); err != nil {
		return fmt.Errorf("invalid sessions config: %w", err)
	}

	if err := c.WorkingHours.validate(); err != nil {
		return fmt.Errorf("invalid working hours config: %w", err)
	}

	return nil
}

func (sc *SessionConfig) validate() error {
	if sc.DurationMinutes <= 0 {
		return fmt.Errorf("session duration must be positive")
	}
	if sc.MinSpacingHours < 0 {
		return fmt.Errorf("minimum session spacing cannot be negative")
	}
	if sc.MaxPerPersonPerWeek <= 0 {
		return fmt.Errorf("maximum sessions per person per week must be positive")
	}
	return nil
}

func (whc *WorkingHoursConfig) validate() error {
	if _, err := time.LoadLocation(whc.Timezone); err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}

	if err := whc.Morning.validate(); err != nil {
		return fmt.Errorf("invalid morning slot: %w", err)
	}

	if err := whc.Afternoon.validate(); err != nil {
		return fmt.Errorf("invalid afternoon slot: %w", err)
	}

	return nil
}

func (tsc *TimeSlotConfig) validate() error {
	if err := tsc.Start.validate(); err != nil {
		return fmt.Errorf("invalid start time: %w", err)
	}

	if err := tsc.End.validate(); err != nil {
		return fmt.Errorf("invalid end time: %w", err)
	}

	start := time.Date(2000, 1, 1, tsc.Start.Hour, tsc.Start.Minute, 0, 0, time.UTC)
	end := time.Date(2000, 1, 1, tsc.End.Hour, tsc.End.Minute, 0, 0, time.UTC)

	if end.Before(start) {
		return fmt.Errorf("end time must be after start time")
	}

	return nil
}

func (tc *TimeConfig) validate() error {
	if tc.Hour < 0 || tc.Hour > 23 {
		return fmt.Errorf("hour must be between 0 and 23")
	}
	if tc.Minute < 0 || tc.Minute > 59 {
		return fmt.Errorf("minute must be between 0 and 59")
	}
	return nil
}
