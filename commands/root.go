package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
