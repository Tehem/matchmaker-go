package commands

import (
	"fmt"
	"log/slog"

	"matchmaker/internal/calendar"
	"matchmaker/internal/config"
	"matchmaker/internal/matching"

	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Create review events in reviewers' calendars",
	Long: `This command will take input from the 'planning.yml' file and create review 
events in reviewers' calendars.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Initialize calendar service
		calendarService, err := calendar.NewService(ctx, "client_secret.json")
		if err != nil {
			return fmt.Errorf("failed to create calendar service: %w", err)
		}

		// Load planning configuration
		matches, err := loadPlanning("planning.yml")
		if err != nil {
			return fmt.Errorf("failed to load planning configuration: %w", err)
		}

		// Create events
		for _, match := range matches {
			event := &calendar.Event{
				Summary:     fmt.Sprintf("%s%s & %s", cfg.Sessions.SessionPrefix, match.Reviewer1.Email, match.Reviewer2.Email),
				Start:       match.TimeSlot.Start,
				End:         match.TimeSlot.End,
				Description: fmt.Sprintf("Code review session\nCommon skills: %v", match.CommonSkills),
				Attendees:   []string{match.Reviewer1.Email, match.Reviewer2.Email},
			}

			if err := calendarService.CreateEvent(ctx, match.Reviewer1.Email, event); err != nil {
				slog.Error("Failed to create event", "reviewer1", match.Reviewer1.Email, "error", err)
				continue
			}

			if err := calendarService.CreateEvent(ctx, match.Reviewer2.Email, event); err != nil {
				slog.Error("Failed to create event", "reviewer2", match.Reviewer2.Email, "error", err)
				continue
			}
		}

		slog.Info("Planning completed successfully", "events", len(matches))
		return nil
	},
}

func init() {
	RootCmd.AddCommand(planCmd)
}

// loadPlanning loads the planning configuration from YAML
func loadPlanning(filename string) ([]matching.Match, error) {
	// TODO: Implement YAML loading
	return nil, fmt.Errorf("not implemented")
}
