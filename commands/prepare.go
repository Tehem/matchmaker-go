package commands

import (
	logrus2 "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	logger "github.com/transcovo/go-chpr-logger"
	"google.golang.org/api/calendar/v3"
	"io/ioutil"
	"matchmaker/libs"
	"matchmaker/libs/gcalendar"
	"matchmaker/util"
	"os"
	"time"
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
			ranges <- GetWorkRange(beginOfWeek, day, 10, 0, 12, 0)
			ranges <- GetWorkRange(beginOfWeek, day, 14, 0, 18, 0)
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

func loadProblem(weekShift int) *libs.Problem {
	people, err := libs.LoadPersons("./persons.yml")
	util.PanicOnError(err, "Can't load people")
	logger.WithField("count", len(people)).Info("People loaded")

	cal, err := gcalendar.GetGoogleCalendarService()
	util.PanicOnError(err, "Can't get gcalendar client")
	logger.Info("Connected to google calendar")

	beginOfWeek := FirstDayOfISOWeek(weekShift)
	logger.WithField("weekFirstDay", beginOfWeek).Info("Planning for week")

	workRanges := ToSlice(GetWeekWorkRanges(beginOfWeek))
	busyTimes := []*libs.BusyTime{}
	for _, person := range people {
		personLogger := logger.WithField("person", person.Email)
		personLogger.Info("Loading busy detail")
		for _, workRange := range workRanges {
			personLogger.WithFields(logrus2.Fields{
				"start": workRange.Start,
				"end":   workRange.End,
			}).Info("Loading busy detail on range")
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
			println(person.Email + ":")
			for _, busyTimePeriod := range busyTimePeriods {
				println("  - " + busyTimePeriod.Start + " -> " + busyTimePeriod.End)
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
	Use:   "prepare",
	Short: "Retrieve available slots for people and parameters for the matching algorithm.",
	Long: `Compute work ranges for the target week, and check free slots for each potential
reviewer and create an output file 'problem.yml'.`,
	Run: func(cmd *cobra.Command, args []string) {

		problem := loadProblem(weekShift)
		yml, _ := problem.ToYaml()
		ioutil.WriteFile("./problem.yml", yml, os.FileMode(0644))
	},
}
