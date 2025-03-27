package commands

import (
	"fmt"
	"log/slog"
	"time"

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
		ctx := cmd.Context()

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Load problem configuration
		people, err := loadProblem("problem.yml")
		if err != nil {
			return fmt.Errorf("failed to load problem configuration: %w", err)
		}

		// Create matcher
		matcherConfig := &matching.Config{
			SessionDuration:     time.Duration(cfg.Sessions.DurationMinutes) * time.Minute,
			MinSessionSpacing:   time.Duration(cfg.Sessions.MinSpacingHours) * time.Hour,
			MaxPerPersonPerWeek: cfg.Sessions.MaxPerPersonPerWeek,
		}
		matcher := matching.NewMatcher(people, matcherConfig)

		// Find matches
		matches, err := matcher.FindMatches()
		if err != nil {
			return fmt.Errorf("failed to find matches: %w", err)
		}

		// Save planning
		if err := savePlanning(matches, cfg.Sessions.SessionPrefix); err != nil {
			return fmt.Errorf("failed to save planning: %w", err)
		}

		slog.Info("Matching completed successfully", "matches", len(matches))
		return nil
	},
}

func init() {
	RootCmd.AddCommand(matchCmd)
}

// loadProblem loads the problem configuration from YAML
func loadProblem(filename string) ([]*matching.Person, error) {
	// TODO: Implement YAML loading
	return nil, fmt.Errorf("not implemented")
}

// savePlanning saves the planning configuration to YAML
func savePlanning(matches []matching.Match, sessionPrefix string) error {
	// TODO: Implement YAML saving
	return fmt.Errorf("not implemented")
}
