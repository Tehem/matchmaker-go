package commands

import (
	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use:   "matchmaker",
	Short: "Matchmaker - A tool for matching and planning reviewers and review slots",
	Long: `Matchmaker takes care of matching and planning of reviewers and review slots 
in people's calendars. It helps organize code review sessions efficiently.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}
