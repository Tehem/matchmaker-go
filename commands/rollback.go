package commands

import (
	"encoding/json"
	"fmt"
	"matchmaker/libs/gcalendar"
	"matchmaker/libs/types"
	"matchmaker/util"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// EventBatch represents a collection of events created in a single plan command run
type EventBatch struct {
	ID        string  `json:"id"`
	CreatedAt string  `json:"created_at"`
	Events    []Event `json:"events"`
}

// Event represents a created calendar event
type Event struct {
	ID        string `json:"id"`
	Summary   string `json:"summary"`
	Organizer string `json:"organizer"`
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback [batch-id]",
	Short: "Rollback events created in a specific batch",
	Long: `Delete all events created in a specific batch.
If no batch ID is provided, the command will prompt for one.
After successful deletion, you can choose to delete the batch file.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var batchID string
		if len(args) > 0 {
			batchID = args[0]
		} else {
			fmt.Print("Enter batch ID: ")
			fmt.Scanln(&batchID)
		}

		// Load the batch file
		batchFile := filepath.Join("batches", fmt.Sprintf("batch-%s.json", batchID))
		data, err := os.ReadFile(batchFile)
		if err != nil {
			util.PanicOnError(err, "Failed to read batch file")
		}

		var batch types.EventBatch
		if err := json.Unmarshal(data, &batch); err != nil {
			util.PanicOnError(err, "Failed to parse batch file")
		}

		logrus.Infof("Found batch created at %s with %d events", batch.CreatedAt, len(batch.Events))

		// Get calendar service
		cal, err := gcalendar.GetGoogleCalendarService()
		util.PanicOnError(err, "Can't get gcalendar client")

		// Track deletion results
		successfulDeletions := 0
		failedDeletions := 0

		// Delete each event
		for _, event := range batch.Events {
			logrus.Infof("Deleting event: %s (ID: %s)", event.Summary, event.ID)
			err := cal.Events.Delete(event.Organizer, event.ID).Do()
			if err != nil {
				logrus.Errorf("Failed to delete event %s: %v", event.ID, err)
				failedDeletions++
			} else {
				logrus.Infof("Successfully deleted event: %s", event.Summary)
				successfulDeletions++
			}
		}

		// Print summary
		logrus.Infof("Rollback summary:")
		logrus.Infof("- Total events: %d", len(batch.Events))
		logrus.Infof("- Successfully deleted: %d", successfulDeletions)
		logrus.Infof("- Failed to delete: %d", failedDeletions)

		// If all deletions were successful, offer to delete the batch file
		if failedDeletions == 0 {
			fmt.Print("\nAll events were successfully deleted. Would you like to delete the batch file? (y/n): ")
			var choice string
			fmt.Scanln(&choice)
			if choice == "y" || choice == "Y" {
				if err := os.Remove(batchFile); err != nil {
					logrus.Warnf("Failed to delete batch file: %v", err)
				} else {
					logrus.Info("Batch file deleted successfully")
				}
			}
		} else {
			logrus.Warn("Batch file was not deleted due to failed event deletions")
		}
	},
}
