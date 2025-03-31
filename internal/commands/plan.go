package commands

import (
	"fmt"

	"matchmaker/internal/calendar"
	"matchmaker/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// PlanCmd represents the plan command
var PlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan review sessions in calendars",
	Long: `This command will take the matches from 'planning.yml' and create calendar events
for each review session.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Read planning configuration
		matches, err := config.LoadPlanning("planning.yml")
		if err != nil {
			return fmt.Errorf("failed to load planning: %w", err)
		}

		// Get session prefix from config
		sessionPrefix := viper.GetString("sessions.sessionPrefix")
		if sessionPrefix == "" {
			return fmt.Errorf("session prefix is required")
		}

		// Get organizer email from config (optional)
		organizerEmail := viper.GetString("organizer_email")

		// Create calendar events for each match
		for _, match := range matches {
			// Create event title with common skills
			title := fmt.Sprintf("%s%s & %s - %s",
				sessionPrefix,
				match.Reviewer1.Email,
				match.Reviewer2.Email,
				match.CommonSkills[0], // Use first common skill as primary focus
			)

			// Create event description
			description := fmt.Sprintf("Code review session between %s and %s\n\nCommon skills: %v",
				match.Reviewer1.Email,
				match.Reviewer2.Email,
				match.CommonSkills,
			)

			// Create attendees list
			attendees := []string{
				match.Reviewer1.Email,
				match.Reviewer2.Email,
			}

			// Create calendar event
			event := &calendar.Event{
				Summary:     title,
				Description: description,
				Start:       match.TimeSlot.Start,
				End:         match.TimeSlot.End,
				Attendees:   attendees,
			}

			// Set organizer and determine which calendar to create the event in
			var calendarEmail string = match.Reviewer1.Email
			if organizerEmail != "" {
				event.OrganizerEmail = organizerEmail
				calendarEmail = organizerEmail
			}

			// Create event in the appropriate calendar
			if err := calendarService.CreateEvent(ctx, calendarEmail, event); err != nil {
				return fmt.Errorf("failed to create calendar event: %w", err)
			}
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(PlanCmd)
}
