package main

import (
	"flag"
	"github.com/Tehem/matchmaker-go/gcalendar"
	"github.com/Tehem/matchmaker-go/match"
	"github.com/Tehem/matchmaker-go/util"
	logrus2 "github.com/sirupsen/logrus"
	logger "github.com/transcovo/go-chpr-logger"
	"google.golang.org/api/calendar/v3"
	"io/ioutil"
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

func GetWorkRange(beginOfWeek time.Time, day int, startHour int, startMinute int, endHour int, endMinute int) *match.Range {
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
	return &match.Range{
		Start: start,
		End:   end,
	}
}

func GetWeekWorkRanges(beginOfWeek time.Time) chan *match.Range {
	ranges := make(chan *match.Range)

	go func() {
		for day := 0; day < 5; day++ {
			ranges <- GetWorkRange(beginOfWeek, day, 10, 0, 12, 0)
			ranges <- GetWorkRange(beginOfWeek, day, 14, 0, 18, 30)
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

func ToSlice(c chan *match.Range) []*match.Range {
	s := make([]*match.Range, 0)
	for r := range c {
		s = append(s, r)
	}
	return s
}

func loadProblem(weekShift int) *match.Problem {
	people, err := match.LoadPersons("./persons.yml")
	util.PanicOnError(err, "Can't load people")
	logger.WithField("count", len(people)).Info("People loaded")

	cal, err := gcalendar.GetGoogleCalendarService()
	util.PanicOnError(err, "Can't get gcalendar client")
	logger.Info("Connected to google calendar")

	beginOfWeek := FirstDayOfISOWeek(weekShift)
	logger.WithField("weekFirstDay", beginOfWeek).Info("Planning for week")

	workRanges := ToSlice(GetWeekWorkRanges(beginOfWeek))
	busyTimes := []*match.BusyTime{}
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
				busyTimes = append(busyTimes, &match.BusyTime{
					Person: person,
					Range: &match.Range{
						Start: parseTime(busyTimePeriod.Start),
						End:   parseTime(busyTimePeriod.End),
					},
				})
			}
		}
	}
	return &match.Problem{
		People:         people,
		WorkRanges:     workRanges,
		BusyTimes:      busyTimes,
		TargetCoverage: 2,
	}
}

// One flag 'week-shift' can be set to plan for an upcoming week instead of next week
// Default = 0 (planning for next week)
// 1 = in two weeks, 2 = in 3 weeks, etc.
func main() {
	weekShiftPtr := flag.Int("week-shift", 0, "Week shift")
	flag.Parse()

	problem := loadProblem(*weekShiftPtr)
	yml, _ := problem.ToYaml()
	ioutil.WriteFile("./problem.yml", yml, os.FileMode(0644))
}
