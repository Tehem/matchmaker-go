package commands

import (
	"fmt"
	"matchmaker/libs/gcalendar"
	"matchmaker/libs/types"
	"matchmaker/util"
	"os"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/calendar/v3"
	"gopkg.in/yaml.v3"
)

func LoadPlan(yml []byte) (*types.Solution, error) {
	var solution *types.Solution
	err := yaml.Unmarshal(yml, &solution)
	if err != nil {
		return nil, err
	}

	return solution, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func choosePlanningFile(args []string) (string, error) {
	// If a file is specified as an argument, use it
	if len(args) > 0 {
		if !fileExists(args[0]) {
			return "", fmt.Errorf("specified file '%s' does not exist", args[0])
		}
		return args[0], nil
	}

	// Check which files exist
	planningExists := fileExists("./planning.yml")
	weeklyPlanningExists := fileExists("./weekly-planning.yml")

	// If neither file exists
	if !planningExists && !weeklyPlanningExists {
		return "", fmt.Errorf("no planning files found (planning.yml or weekly-planning.yml)")
	}

	// If only one file exists, use it
	if planningExists && !weeklyPlanningExists {
		return "./planning.yml", nil
	}
	if !planningExists && weeklyPlanningExists {
		return "./weekly-planning.yml", nil
	}

	// Both files exist, ask user which one to use
	fmt.Println("Multiple planning files found. Please choose which one to use:")
	fmt.Println("1) planning.yml")
	fmt.Println("2) weekly-planning.yml")

	var choice string
	fmt.Print("Enter choice (1 or 2): ")
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		return "./planning.yml", nil
	case "2":
		return "./weekly-planning.yml", nil
	default:
		return "", fmt.Errorf("invalid choice: %s", choice)
	}
}

func init() {
	rootCmd.AddCommand(planCmd)
}

var planCmd = &cobra.Command{
	Use:   "plan [file]",
	Short: "Create events in people's calendars.",
	Long: `Take input from a planning file (planning.yml or weekly-planning.yml) and create session events in people's Google Calendar.
If no file is specified, the command will:
- Use planning.yml if it's the only file present
- Use weekly-planning.yml if it's the only file present
- Ask which file to use if both are present`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		planningFile, err := choosePlanningFile(args)
		util.PanicOnError(err, "Failed to determine planning file")

		util.LogInfo("Using planning file", map[string]interface{}{
			"file": planningFile,
		})

		yml, err := os.ReadFile(planningFile)
		util.PanicOnError(err, "Can't read planning file")

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
