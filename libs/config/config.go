package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

const (
	// Config keys
	WorkingHoursMorningStartHour     = "workingHours.morning.start.hour"
	WorkingHoursMorningStartMinute   = "workingHours.morning.start.minute"
	WorkingHoursMorningEndHour       = "workingHours.morning.end.hour"
	WorkingHoursMorningEndMinute     = "workingHours.morning.end.minute"
	WorkingHoursAfternoonStartHour   = "workingHours.afternoon.start.hour"
	WorkingHoursAfternoonStartMinute = "workingHours.afternoon.start.minute"
	WorkingHoursAfternoonEndHour     = "workingHours.afternoon.end.hour"
	WorkingHoursAfternoonEndMinute   = "workingHours.afternoon.end.minute"
	WorkingHoursTimezone             = "workingHours.timezone"
	SessionDurationMinutes           = "sessions.sessionDurationMinutes"
	MinSessionSpacingHours           = "sessions.minSessionSpacingHours"
	MaxSessionsPerPersonPerWeek      = "sessions.maxPerPersonPerWeek"
	SessionPrefix                    = "sessions.sessionPrefix"
)

// WorkHoursConfig represents the configuration for work hours
type WorkHoursConfig struct {
	StartHour   int
	StartMinute int
	EndHour     int
	EndMinute   int
}

// GetWorkHoursConfig retrieves work hours configuration for a specific period
func GetWorkHoursConfig(period string) (*WorkHoursConfig, error) {
	prefix := fmt.Sprintf("workingHours.%s", period)

	startHour := viper.GetInt(fmt.Sprintf("%s.start.hour", prefix))
	startMinute := viper.GetInt(fmt.Sprintf("%s.start.minute", prefix))
	endHour := viper.GetInt(fmt.Sprintf("%s.end.hour", prefix))
	endMinute := viper.GetInt(fmt.Sprintf("%s.end.minute", prefix))

	if err := validateTimeRange(startHour, startMinute, endHour, endMinute); err != nil {
		return nil, fmt.Errorf("invalid %s configuration: %w", period, err)
	}

	return &WorkHoursConfig{
		StartHour:   startHour,
		StartMinute: startMinute,
		EndHour:     endHour,
		EndMinute:   endMinute,
	}, nil
}

// GetTimezone returns the configured timezone
func GetTimezone() string {
	return viper.GetString(WorkingHoursTimezone)
}

// GetSessionDuration returns the configured session duration
func GetSessionDuration() time.Duration {
	return time.Duration(viper.GetInt(SessionDurationMinutes)) * time.Minute
}

// GetMinSessionSpacing returns the minimum spacing between sessions
func GetMinSessionSpacing() time.Duration {
	return time.Duration(viper.GetInt(MinSessionSpacingHours)) * time.Hour
}

// GetMaxSessionsPerPersonPerWeek returns the maximum number of sessions a person can have per week
func GetMaxSessionsPerPersonPerWeek() int {
	return viper.GetInt(MaxSessionsPerPersonPerWeek)
}

// GetSessionPrefix returns the prefix used for session event titles
func GetSessionPrefix() string {
	return viper.GetString(SessionPrefix)
}

// validateTimeRange checks if the time range is valid
func validateTimeRange(startHour, startMinute, endHour, endMinute int) error {
	if startHour < 0 || startHour >= 24 || endHour < 0 || endHour >= 24 {
		return fmt.Errorf("invalid hour: must be between 0 and 23")
	}

	if startMinute < 0 || startMinute >= 60 || endMinute < 0 || endMinute >= 60 {
		return fmt.Errorf("invalid minute: must be between 0 and 59")
	}

	startTotal := startHour*60 + startMinute
	endTotal := endHour*60 + endMinute

	if endTotal <= startTotal {
		return fmt.Errorf("end time must be after start time")
	}

	return nil
}
