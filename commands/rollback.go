package commands

import (
	"encoding/json"
	"fmt"
	"matchmaker/libs/gcalendar"
	"matchmaker/libs/types"
	"matchmaker/libs/util"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/api/calendar/v3"
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

// BatchInfo represents information about a batch file
type BatchInfo struct {
	ID         string
	CreatedAt  time.Time
	Filename   string
	EventCount int
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback [batch-id]",
	Short: "Rollback events created in a specific batch",
	Long: `Delete all events created in a specific batch.
If no batch ID is provided, you will be shown a list of available batches to choose from.
After successful deletion, you can choose to delete the batch file.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		batchID := getBatchID(args)
		batch := loadBatch(batchID)
		rollbackBatch(batch, batchID)
	},
}

// getBatchID returns the batch ID either from args or by prompting the user
func getBatchID(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return promptForBatchID()
}

// promptForBatchID shows a list of available batches and prompts the user to select one
func promptForBatchID() string {
	batches := listAvailableBatches()
	if len(batches) == 0 {
		util.PanicOnError(fmt.Errorf("no batch files found"), "No batches available")
	}

	displayBatches(batches)
	return selectBatch(batches)
}

// listAvailableBatches returns a list of available batch files sorted by creation date
func listAvailableBatches() []BatchInfo {
	entries, err := os.ReadDir("batches")
	if err != nil {
		util.PanicOnError(err, "Failed to read batches directory")
	}

	var batches []BatchInfo
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "batch-") && strings.HasSuffix(entry.Name(), ".json") {
			batch, err := loadBatchInfo(entry.Name())
			if err != nil {
				logrus.Warnf("Skipping batch file %s: %v", entry.Name(), err)
				continue
			}
			batches = append(batches, batch)
		}
	}

	sort.Slice(batches, func(i, j int) bool {
		return batches[i].CreatedAt.After(batches[j].CreatedAt)
	})

	return batches
}

// loadBatchInfo loads batch information from a file
func loadBatchInfo(filename string) (BatchInfo, error) {
	data, err := os.ReadFile(filepath.Join("batches", filename))
	if err != nil {
		return BatchInfo{}, fmt.Errorf("failed to read file: %w", err)
	}

	var batch types.EventBatch
	if err := json.Unmarshal(data, &batch); err != nil {
		return BatchInfo{}, fmt.Errorf("failed to parse file: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, batch.CreatedAt)
	if err != nil {
		return BatchInfo{}, fmt.Errorf("failed to parse creation date: %w", err)
	}

	return BatchInfo{
		ID:         strings.TrimPrefix(strings.TrimSuffix(filename, ".json"), "batch-"),
		CreatedAt:  createdAt,
		Filename:   filename,
		EventCount: len(batch.Events),
	}, nil
}

// displayBatches shows a numbered list of available batches
func displayBatches(batches []BatchInfo) {
	fmt.Println("\nAvailable batches:")
	for i, batch := range batches {
		fmt.Printf("%d. Batch ID: %s (created at %s, %d events)\n",
			i+1,
			batch.ID,
			batch.CreatedAt.Format("2006-01-02 15:04:05"),
			batch.EventCount)
	}
}

// selectBatch prompts the user to select a batch and returns its ID
func selectBatch(batches []BatchInfo) string {
	fmt.Print("\nEnter the number of the batch to rollback: ")
	var choice int
	fmt.Scanln(&choice)
	if choice < 1 || choice > len(batches) {
		util.PanicOnError(fmt.Errorf("invalid choice"), "Please select a valid batch number")
	}
	return batches[choice-1].ID
}

// loadBatch loads a batch file by ID
func loadBatch(batchID string) *types.EventBatch {
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
	return &batch
}

// rollbackBatch deletes all events in a batch and handles the batch file
func rollbackBatch(batch *types.EventBatch, batchID string) {
	cal, err := gcalendar.GetCalendarService()
	util.PanicOnError(err, "Can't get gcalendar client")

	successfulDeletions, failedDeletions := deleteEvents(cal, batch.Events)
	printRollbackSummary(len(batch.Events), successfulDeletions, failedDeletions)

	if failedDeletions == 0 {
		handleBatchFileDeletion(batchID)
	} else {
		logrus.Warn("Batch file was not deleted due to failed event deletions")
	}
}

// deleteEvents deletes all events in a batch and returns the results
func deleteEvents(cal *calendar.Service, events []types.Event) (int, int) {
	successfulDeletions := 0
	failedDeletions := 0

	for _, event := range events {
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

	return successfulDeletions, failedDeletions
}

// printRollbackSummary displays the results of the rollback operation
func printRollbackSummary(totalEvents, successfulDeletions, failedDeletions int) {
	logrus.Infof("Rollback summary:")
	logrus.Infof("- Total events: %d", totalEvents)
	logrus.Infof("- Successfully deleted: %d", successfulDeletions)
	logrus.Infof("- Failed to delete: %d", failedDeletions)
}

// handleBatchFileDeletion prompts the user to delete the batch file if all events were deleted successfully
func handleBatchFileDeletion(batchID string) {
	fmt.Print("\nAll events were successfully deleted. Would you like to delete the batch file? (y/n): ")
	var choice string
	fmt.Scanln(&choice)
	if choice == "y" || choice == "Y" {
		batchFile := filepath.Join("batches", fmt.Sprintf("batch-%s.json", batchID))
		if err := os.Remove(batchFile); err != nil {
			logrus.Warnf("Failed to delete batch file: %v", err)
		} else {
			logrus.Info("Batch file deleted successfully")
		}
	}
}
