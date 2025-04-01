package commands

import (
	"matchmaker/util"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "matchmaker",
	Short: "A Google Calendar matchmaking CLI for pairings",
	Long: `Matchmaker is a simple CLI to match people from a group to spend 
time together by putting common appointments in their Google Calendar when 
they are both not busy.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		util.LogError(err, "Failed to execute command")
		os.Exit(1)
	}
}
