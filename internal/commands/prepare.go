package commands

import (
	"fmt"
	"log/slog"
	"time"

	"matchmaker/internal/calendar"
	"matchmaker/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	weekShift       int
	calendarService calendar.CalendarService
	timeNow         = time.Now // Function to get current time, can be overridden in tests
)

// PrepareCmd represents the prepare command
var PrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare the matching process by computing work ranges and checking free slots",
	Long: `This command will compute work ranges for the target week, and check free slots 
for each potential reviewer and create an output file 'problem.yml'.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Initialize calendar service if not already set
		if calendarService == nil {
			var err error
			calendarService, err = calendar.NewCalendarServiceFromToken(ctx)
			if err != nil {
				return fmt.Errorf("failed to create calendar service: %w", err)
			}
		}

		// Calculate target week
		targetWeek := calculateTargetWeek(weekShift)
		slog.Info("Preparing for week", "firstDay", targetWeek.Format("2006-01-02"), "lastDay", targetWeek.AddDate(0, 0, 6).Format("2006-01-02"))

		// Load people configuration
		people, err := config.LoadPeople("persons.yml")
		if err != nil {
			return fmt.Errorf("failed to load people configuration: %w", err)
		}
		slog.Info("Loaded people configuration", "people", len(people))

		// Get free slots for each person
		for _, person := range people {
			// Get busy slots for the person
			busySlots, err := calendarService.GetBusySlots(ctx, person.Email, targetWeek, targetWeek.AddDate(0, 0, 7))
			if err != nil {
				slog.Error("Failed to get busy slots", "person", person.Email, "error", err)
				continue
			}
			slog.Info("Busy slots", "person", person.Email, "busySlots", len(busySlots))
			outputSlots("Busy", busySlots)

			// Convert busy slots to events
			busyEvents := make([]*calendar.Event, len(busySlots))
			for i, slot := range busySlots {
				busyEvents[i] = &calendar.Event{
					Start: slot.Start,
					End:   slot.End,
				}
			}

			// Get free slots using the busy events
			slots, err := calendarService.GetFreeSlots(ctx, person.Email, targetWeek, targetWeek.AddDate(0, 0, 7), busyEvents)
			if err != nil {
				slog.Error("Failed to get free slots", "person", person.Email, "error", err)
				continue
			}
			person.FreeSlots = slots
			slog.Info("Free slots", "person", person.Email, "freeSlots", len(slots))
			outputSlots("Free", slots)
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
	RootCmd.AddCommand(PrepareCmd)

	PrepareCmd.Flags().IntVarP(&weekShift, "week-shift", "w", 0, "Number of weeks to shift from current week (0 = next week)")
}

func outputSlots(prefix string, slots []calendar.TimeSlot) {
	// use correct timezone
	loc, err := time.LoadLocation(viper.GetString("workingHours.timezone"))
	if err != nil {
		slog.Error("Failed to load location", "error", err)
		return
	}
	for _, slot := range slots {
		slog.Info(prefix, "from", slot.Start.In(loc).Format("2006-01-02 15:04"), "to", slot.End.In(loc).Format("2006-01-02 15:04"))
	}
}

// calculateTargetWeek calculates the target week based on the current time and week shift
func calculateTargetWeek(shift int) time.Time {
	return calculateTargetWeekFromDate(timeNow(), shift)
}

// calculateTargetWeekFromDate calculates the target week based on a given date and week shift
func calculateTargetWeekFromDate(now time.Time, shift int) time.Time {
	// Calculate days until next Monday
	daysUntilMonday := (8 - int(now.Weekday())) % 7

	// Add days to get to next Monday
	nextMonday := now.AddDate(0, 0, daysUntilMonday)

	// Add weeks based on shift
	nextMonday = nextMonday.AddDate(0, 0, shift*7)

	// Return date at midnight UTC
	return time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 0, 0, 0, 0, time.UTC)
}
