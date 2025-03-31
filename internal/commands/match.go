package commands

import (
	"fmt"

	"matchmaker/internal/config"
	"matchmaker/internal/matching"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// MatchCmd represents the match command
var MatchCmd = &cobra.Command{
	Use:   "match",
	Short: "Match reviewers with review slots",
	Long: `This command will match reviewers with review slots based on their availability
and skills. It will create a new file 'planning.yml' with the matches.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read problem configuration
		people, err := config.LoadProblem("problem.yml")
		if err != nil {
			return fmt.Errorf("failed to load problem: %w", err)
		}

		// Check if we have enough people to create matches
		if len(people) < 2 {
			return fmt.Errorf("need at least 2 people to create matches")
		}

		// Create matcher configuration from Viper
		matcherConfig := &matching.Config{
			SessionDuration:     viper.GetDuration("session.duration"),
			MinSessionSpacing:   viper.GetDuration("session.min_spacing"),
			MaxPerPersonPerWeek: viper.GetInt("session.max_per_person_per_week"),
		}

		// Create matcher and find matches
		matcher := matching.NewMatcher(people, matcherConfig)
		matches, err := matcher.FindMatches()
		if err != nil {
			return fmt.Errorf("failed to find matches: %w", err)
		}

		// Save matches to planning.yml
		if err := config.SavePlanning(matches, "planning.yml"); err != nil {
			return fmt.Errorf("failed to save planning: %w", err)
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(MatchCmd)
}
