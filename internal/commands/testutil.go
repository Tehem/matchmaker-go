package commands

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ResetRootCmdForTest resets the root command to its initial state for testing
func ResetRootCmdForTest() {
	RootCmd = &cobra.Command{
		Use:   "matchmaker",
		Short: "Matchmaker - A tool for matching and planning reviewers and review slots",
		Long: `Matchmaker takes care of matching and planning of reviewers and review slots 
in people's calendars. It helps organize code review sessions efficiently.`,
	}
	// Re-add all commands
	RootCmd.AddCommand(PrepareCmd)
	RootCmd.AddCommand(MatchCmd)
	RootCmd.AddCommand(PlanCmd)
	RootCmd.AddCommand(TokenCmd)
}

// SetupTestConfig configures Viper with default test values
func SetupTestConfig() {
	viper.Set("sessions.minSpacing", 24*time.Hour)
	viper.Set("sessions.maxPerPersonPerWeek", 2)
	viper.Set("sessions.sessionPrefix", "Code Review: ")
	viper.Set("organizer_email", "test@example.com")

	// Set working hours with morning and afternoon slots
	viper.Set("workingHours.timezone", "UTC")
	viper.Set("workingHours.morning.start", "09:00")
	viper.Set("workingHours.morning.end", "12:00")
	viper.Set("workingHours.afternoon.start", "13:00")
	viper.Set("workingHours.afternoon.end", "17:00")
}
