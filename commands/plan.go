package commands

import (
	"matchmaker/libs"
	"matchmaker/libs/gcalendar"
	"matchmaker/util"
	"os"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/calendar/v3"
	"gopkg.in/yaml.v3"
)

func LoadPlan(yml []byte) (*libs.Solution, error) {
	var solution *libs.Solution
	err := yaml.Unmarshal(yml, &solution)
	if err != nil {
		return nil, err
	}

	return solution, nil
}

func init() {
	rootCmd.AddCommand(planCmd)
}

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Create events in people's calendars.",
	Long:  `Take input from the 'planning.yml' file and create session events in people's Google Calendar.`,
	Run: func(cmd *cobra.Command, args []string) {
		yml, err := os.ReadFile("./planning.yml")
		util.PanicOnError(err, "Can't yml problem description")

		cal, err := gcalendar.GetGoogleCalendarService()
		util.PanicOnError(err, "Can't get gcalendar client")

		solution, err := LoadPlan(yml)
		util.PanicOnError(err, "Can't get solution from planning file")

		// calendar owner
		masterEmail := viper.GetString("organizerEmail")

		for _, session := range solution.Sessions {
			attendees := []*calendar.EventAttendee{}

			for _, person := range session.Reviewers.People {
				attendees = append(attendees, &calendar.EventAttendee{
					Email: person.Email,
				})
			}

			// take first attendee as organizer
			organizer := attendees[0].Email

			// add master email as optional, and use it as organizer by default
			if masterEmail != "" {
				organizer = masterEmail
				attendees = append(attendees, &calendar.EventAttendee{
					Email:    masterEmail,
					Optional: true,
				})
			}

			event := &calendar.Event{
				Start: &calendar.EventDateTime{
					DateTime: gcalendar.FormatTime(session.Range.Start),
					TimeZone: viper.GetString("workingHours.timezone"),
				},
				End: &calendar.EventDateTime{
					DateTime: gcalendar.FormatTime(session.Range.End),
					TimeZone: viper.GetString("workingHours.timezone"),
				},
				Summary:         session.GetDisplayName(),
				Attendees:       attendees,
				GuestsCanModify: true,
				ConferenceData: &calendar.ConferenceData{
					CreateRequest: &calendar.CreateConferenceRequest{
						RequestId: uuid.New().String(),
						ConferenceSolutionKey: &calendar.ConferenceSolutionKey{
							Type: "hangoutsMeet",
						},
						Status: &calendar.ConferenceRequestStatus{
							StatusCode: "success",
						},
					},
				},
			}

			_, err := cal.Events.Insert(organizer, event).ConferenceDataVersion(1).Do()
			util.PanicOnError(err, "Can't create event")
			logrus.Info("✔ " + session.GetDisplayName())
		}
	},
}
