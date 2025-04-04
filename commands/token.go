package commands

import (
	"matchmaker/libs/gcalendar"
	"matchmaker/util"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tokenCmd)
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Retrieve a Google Calendar API token.",
	Long:  `Authorize the app to access your Google Agenda and get an auth token for Google Calendar API.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the Google Calendar client
		cal, err := gcalendar.NewGCalendar()
		if err != nil {
			util.LogError(err, "Unable to retrieve Calendar client")
			return
		}

		events, err := cal.GetNextEvents(time.Now(), 10)
		if err != nil {
			util.LogError(err, "Unable to retrieve next ten of the user's events")
			return
		}

		util.LogInfo("Upcoming events", nil)
		if len(events) == 0 {
			util.LogInfo("No upcoming events found", nil)
		} else {
			for _, item := range events {
				date := item.Start.DateTime
				if date == "" {
					date = item.Start.Date
				}
				util.LogInfo("Event", map[string]interface{}{
					"summary": item.Summary,
					"date":    date,
				})
			}
		}
	},
}
