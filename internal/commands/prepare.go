package commands

import (
	"fmt"
	"log/slog"
	"time"

	"matchmaker/internal/calendar"
	"matchmaker/internal/config"

	"github.com/spf13/cobra"
)

var (
	weekShift int
)

// prepareCmd represents the prepare command
var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare the matching process by computing work ranges and checking free slots",
	Long: `This command will compute work ranges for the target week, and check free slots 
for each potential reviewer and create an output file 'problem.yml'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Load configuration
		if _, err := config.LoadConfig(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Initialize calendar service
		calendarService, err := calendar.NewService(ctx, "client_secret.json")
		if err != nil {
			return fmt.Errorf("failed to create calendar service: %w", err)
		}

		// Calculate target week
		targetWeek := calculateTargetWeek(weekShift)
		slog.Info("Preparing for week", "week", targetWeek.Format("2006-01-02"))

		// Load people configuration
		people, err := config.LoadPeople("persons.yml")
		if err != nil {
			return fmt.Errorf("failed to load people configuration: %w", err)
		}

		// Get free slots for each person
		for _, person := range people {
			slots, err := calendarService.GetFreeSlots(ctx, person.Email, targetWeek, targetWeek.AddDate(0, 0, 7))
			if err != nil {
				slog.Error("Failed to get free slots", "person", person.Email, "error", err)
				continue
			}
			person.FreeSlots = slots
		}

		// Save problem configuration
		if err := config.SaveProblem(people, targetWeek, "problem.yml"); err != nil {
			return fmt.Errorf("failed to save problem configuration: %w", err)
		}

		slog.Info("Preparation completed successfully")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(prepareCmd)

	prepareCmd.Flags().IntVarP(&weekShift, "week-shift", "w", 0, "Number of weeks to shift from current week (0 = next week)")
}

// calculateTargetWeek calculates the target week based on the shift
func calculateTargetWeek(shift int) time.Time {
	now := time.Now()
	// Find next Monday
	daysUntilMonday := (8 - int(now.Weekday())) % 7
	nextMonday := now.AddDate(0, 0, daysUntilMonday)
	// Add week shift
	return nextMonday.AddDate(0, 0, shift*7)
}
