package commands

import (
	"fmt"
	"log/slog"

	"matchmaker/internal/config"
	"matchmaker/internal/matching"

	"github.com/spf13/cobra"
)

// matchCmd represents the match command
var matchCmd = &cobra.Command{
	Use:   "match",
	Short: "Match reviewers together in review slots",
	Long: `This command will take input from the 'problem.yml' file and match reviewers 
together in review slots for the target week. The output is a 'planning.yml' file 
with reviewers couples and planned slots.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.LoadConfig("configs/config.json")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Load problem configuration
		people, err := config.LoadProblem("problem.yml")
		if err != nil {
			return fmt.Errorf("failed to load problem configuration: %w", err)
		}

		// Create matcher
		matcherConfig := &matching.Config{
			SessionDuration:     cfg.Sessions.Duration,
			MinSessionSpacing:   cfg.Sessions.MinSpacing,
			MaxPerPersonPerWeek: cfg.Sessions.MaxPerPersonPerWeek,
		}
		matcher := matching.NewMatcher(people, matcherConfig)

		// Find matches
		matches, err := matcher.FindMatches()
		if err != nil {
			return fmt.Errorf("failed to find matches: %w", err)
		}

		// Save planning
		if err := config.SavePlanning(matches, "planning.yml"); err != nil {
			return fmt.Errorf("failed to save planning: %w", err)
		}

		slog.Info("Matching completed successfully", "matches", len(matches))
		return nil
	},
}

func init() {
	RootCmd.AddCommand(matchCmd)
}
