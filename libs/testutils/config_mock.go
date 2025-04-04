// Package testutils provides utilities for testing, including configuration mocking.
// These utilities help create consistent and maintainable tests across the codebase.
package testutils

import (
	"github.com/spf13/viper"
)

// ConfigMock provides a way to mock configuration for testing purposes.
// It manages the state of viper configuration and provides methods to:
// - Set up default configuration values
// - Modify specific configuration values
// - Reset configuration to default values
// - Restore original configuration
//
// Usage example:
//
//	configMock := NewConfigMock()
//	configMock.SetupWorkHours()
//	defer configMock.Restore()
//	// Run tests...
type ConfigMock struct {
	originalViper *viper.Viper
}

// NewConfigMock creates a new configuration mock instance.
// It stores the current viper configuration to allow restoration later.
func NewConfigMock() *ConfigMock {
	return &ConfigMock{
		originalViper: viper.GetViper(),
	}
}

// SetupWorkHours sets up the default work hours configuration.
// Default values:
// - Morning: 9:00 - 12:00
// - Afternoon: 13:00 - 17:00
func (m *ConfigMock) SetupWorkHours() {
	viper.SetDefault("workingHours.morning.start.hour", 9)
	viper.SetDefault("workingHours.morning.start.minute", 0)
	viper.SetDefault("workingHours.morning.end.hour", 12)
	viper.SetDefault("workingHours.morning.end.minute", 0)
	viper.SetDefault("workingHours.afternoon.start.hour", 13)
	viper.SetDefault("workingHours.afternoon.start.minute", 0)
	viper.SetDefault("workingHours.afternoon.end.hour", 17)
	viper.SetDefault("workingHours.afternoon.end.minute", 0)
	viper.SetDefault("sessions.sessionDurationMinutes", 60)
}

// Reset resets the configuration to default values.
// This is useful between test cases to ensure a clean state.
func (m *ConfigMock) Reset() {
	viper.Reset()
	m.SetupWorkHours()
}

// SetMorningHours sets the morning work hours configuration.
// Parameters:
// - startHour: Start hour (0-23)
// - startMinute: Start minute (0-59)
// - endHour: End hour (0-23)
// - endMinute: End minute (0-59)
func (m *ConfigMock) SetMorningHours(startHour, startMinute, endHour, endMinute int) {
	viper.Set("workingHours.morning.start.hour", startHour)
	viper.Set("workingHours.morning.start.minute", startMinute)
	viper.Set("workingHours.morning.end.hour", endHour)
	viper.Set("workingHours.morning.end.minute", endMinute)
}

// SetAfternoonHours sets the afternoon work hours configuration.
// Parameters:
// - startHour: Start hour (0-23)
// - startMinute: Start minute (0-59)
// - endHour: End hour (0-23)
// - endMinute: End minute (0-59)
func (m *ConfigMock) SetAfternoonHours(startHour, startMinute, endHour, endMinute int) {
	viper.Set("workingHours.afternoon.start.hour", startHour)
	viper.Set("workingHours.afternoon.start.minute", startMinute)
	viper.Set("workingHours.afternoon.end.hour", endHour)
	viper.Set("workingHours.afternoon.end.minute", endMinute)
}

// Restore restores the original configuration that was in place
// when the ConfigMock was created. This should typically be called
// using defer after creating the mock:
//
//	configMock := NewConfigMock()
//	defer configMock.Restore()
func (m *ConfigMock) Restore() {
	viper.SetDefault("workingHours.morning.start.hour", m.originalViper.GetInt("workingHours.morning.start.hour"))
	viper.SetDefault("workingHours.morning.start.minute", m.originalViper.GetInt("workingHours.morning.start.minute"))
	viper.SetDefault("workingHours.morning.end.hour", m.originalViper.GetInt("workingHours.morning.end.hour"))
	viper.SetDefault("workingHours.morning.end.minute", m.originalViper.GetInt("workingHours.morning.end.minute"))
	viper.SetDefault("workingHours.afternoon.start.hour", m.originalViper.GetInt("workingHours.afternoon.start.hour"))
	viper.SetDefault("workingHours.afternoon.start.minute", m.originalViper.GetInt("workingHours.afternoon.start.minute"))
	viper.SetDefault("workingHours.afternoon.end.hour", m.originalViper.GetInt("workingHours.afternoon.end.hour"))
	viper.SetDefault("workingHours.afternoon.end.minute", m.originalViper.GetInt("workingHours.afternoon.end.minute"))
}
