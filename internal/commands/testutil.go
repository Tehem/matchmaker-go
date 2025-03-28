package commands

import (
	"github.com/spf13/cobra"
)

// resetRootCmdForTest resets the root command to its initial state for testing
func resetRootCmdForTest() {
	RootCmd = &cobra.Command{
		Use:   "matchmaker",
		Short: "Matchmaker - A tool for matching and planning reviewers and review slots",
		Long: `Matchmaker takes care of matching and planning of reviewers and review slots 
in people's calendars. It helps organize code review sessions efficiently.`,
	}
	// Re-add all commands
	RootCmd.AddCommand(prepareCmd)
	RootCmd.AddCommand(matchCmd)
	RootCmd.AddCommand(planCmd)
	RootCmd.AddCommand(tokenCmd)
}
