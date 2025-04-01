package commands

import (
	"matchmaker/libs"
	"matchmaker/libs/gcalendar"
	"matchmaker/util"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/calendar/v3"
)

func FirstDayOfISOWeek(weekShift int) time.Time {
	date := time.Now()
	date = time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), 0, 0, 0, date.Location())
	date = date.AddDate(0, 0, 7*weekShift)

	// iterate to Monday
	for !(date.Weekday() == time.Monday && date.Hour() == 0) {
		date = date.Add(time.Hour)
	}

	date = date.AddDate(0, 0, 0)

	return date
}

func GetWorkRange(beginOfWeek time.Time, day int, startHour int, startMinute int, endHour int, endMinute int) *libs.Range {
	start := time.Date(
		beginOfWeek.Year(),
		beginOfWeek.Month(),
		beginOfWeek.Day()+day,
		startHour,
		startMinute,
		0,
		0,
		beginOfWeek.Location(),
	)
	end := time.Date(
		beginOfWeek.Year(),
		beginOfWeek.Month(),
		beginOfWeek.Day()+day,
		endHour,
		endMinute,
		0,
		0,
		beginOfWeek.Location(),
	)
	return &libs.Range{
		Start: start,
		End:   end,
	}
}

func GetWeekWorkRanges(beginOfWeek time.Time) chan *libs.Range {
	ranges := make(chan *libs.Range)

	go func() {
		for day := 0; day < 5; day++ {
			ranges <- GetWorkRange(beginOfWeek, day,
				viper.GetInt("workingHours.morning.start.hour"),
				viper.GetInt("workingHours.morning.start.minute"),
				viper.GetInt("workingHours.morning.end.hour"),
				viper.GetInt("workingHours.morning.end.minute"))
			ranges <- GetWorkRange(beginOfWeek, day,
				viper.GetInt("workingHours.afternoon.start.hour"),
				viper.GetInt("workingHours.afternoon.start.minute"),
				viper.GetInt("workingHours.afternoon.end.hour"),
				viper.GetInt("workingHours.afternoon.end.minute"))
		}
		close(ranges)
	}()

	return ranges
}

func parseTime(dateStr string) time.Time {
	result, err := time.Parse(time.RFC3339, dateStr)
	util.PanicOnError(err, "Impossible to parse date "+dateStr)
	return result
}

func ToSlice(c chan *libs.Range) []*libs.Range {
	s := make([]*libs.Range, 0)
	for r := range c {
		s = append(s, r)
	}
	return s
}

func loadProblem(weekShift int, groupFile string) *libs.Problem {
	groupPath := filepath.Join("groups", groupFile)
	people, err := libs.LoadPersons(groupPath)
	util.PanicOnError(err, "Cannot load people file")
	util.LogInfo("People file loaded", map[string]interface{}{
		"count": len(people),
		"file":  groupPath,
	})

	cal, err := gcalendar.GetGoogleCalendarService()
	util.PanicOnError(err, "Can't get Google Calendar client")
	util.LogInfo("Connected to google calendar", nil)

	beginOfWeek := FirstDayOfISOWeek(weekShift)
	util.LogInfo("Planning for week", map[string]interface{}{
		"weekFirstDay": beginOfWeek,
	})

	workRanges := ToSlice(GetWeekWorkRanges(beginOfWeek))
	busyTimes := []*libs.BusyTime{}
	for _, person := range people {
		if person.MaxSessionsPerWeek == 0 {
			continue
		}
		util.LogInfo("Loading busy detail", map[string]interface{}{
			"person": person.Email,
		})
		for _, workRange := range workRanges {
			util.LogInfo("Loading busy detail on range", map[string]interface{}{
				"start": workRange.Start,
				"end":   workRange.End,
			})
			result, err := cal.Freebusy.Query(&calendar.FreeBusyRequest{
				TimeMin: gcalendar.FormatTime(workRange.Start),
				TimeMax: gcalendar.FormatTime(workRange.End),
				Items: []*calendar.FreeBusyRequestItem{
					{
						Id: person.Email,
					},
				},
			}).Do()
			util.PanicOnError(err, "Can't retrieve free/busy data for "+person.Email)
			busyTimePeriods := result.Calendars[person.Email].Busy
			util.LogInfo("Person busy times", map[string]interface{}{
				"person": person.Email,
			})
			for _, busyTimePeriod := range busyTimePeriods {
				util.LogInfo("Busy time period", map[string]interface{}{
					"start": busyTimePeriod.Start,
					"end":   busyTimePeriod.End,
				})
				busyTimes = append(busyTimes, &libs.BusyTime{
					Person: person,
					Range: &libs.Range{
						Start: parseTime(busyTimePeriod.Start),
						End:   parseTime(busyTimePeriod.End),
					},
				})
			}
		}
	}
	return &libs.Problem{
		People:         people,
		WorkRanges:     workRanges,
		BusyTimes:      busyTimes,
		TargetCoverage: 2,
	}
}

func init() {
	prepareCmd.Flags().IntVarP(&weekShift, "week-shift", "w", 0, `define a week shift to plan for an upcoming 
week instead of next week. Default value (0) is next week, and 1 is the week after, etc.`)

	rootCmd.AddCommand(prepareCmd)
}

// One flag 'week-shift' can be set to plan for an upcoming week instead of next week
// Default = 0 (planning for next week)
// 1 = in two weeks, 2 = in 3 weeks, etc.
var weekShift int

var prepareCmd = &cobra.Command{
	Use:   "prepare [group-file]",
	Short: "Retrieve available slots for a group of people and parameters for the matching algorithm.",
	Long: `Compute work ranges for the target week, and check free slots for each potential
reviewer in a group and create an output file 'problem.yml'.

The group-file parameter specifies which group file to use from the groups directory.
You can create multiple group files (e.g., teams.yml, projects.yml) to manage different sets of people.
If no group file is specified, 'group.yml' will be used by default.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupFile := "group.yml"
		if len(args) > 0 {
			groupFile = args[0]
		}
		problem := loadProblem(weekShift, groupFile)
		yml, _ := problem.ToYaml()
		err := os.WriteFile("./problem.yml", yml, os.FileMode(0644))
		util.PanicOnError(err, "Can't yml problem file")
	},
}
