package commands

import (
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	logger "github.com/transcovo/go-chpr-logger"
	"google.golang.org/api/calendar/v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"matchmaker/libs"
	"matchmaker/libs/gcalendar"
	"matchmaker/util"
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
		yml, err := ioutil.ReadFile("./planning.yml")
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

			_, err := cal.Events.Insert(organizer, &calendar.Event{
				Start: &calendar.EventDateTime{
					DateTime: gcalendar.FormatTime(session.Range.Start),
					TimeZone: viper.GetString("workingHours.timezone"),
				},
				End: &calendar.EventDateTime{
					DateTime: gcalendar.FormatTime(session.Range.End),
					TimeZone: viper.GetString("workingHours.timezone"),
				},
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
				Summary:         session.GetDisplayName(),
				Attendees:       attendees,
				GuestsCanModify: true,
			}).ConferenceDataVersion(1).Do()
			util.PanicOnError(err, "Can't create event")
			logger.Info("✔ " + session.GetDisplayName())
		}
	},
}
