package timeutil

import (
	"matchmaker/libs"
	"time"

	"github.com/spf13/viper"
)

// FirstDayOfISOWeek returns the first day (Monday) of the ISO week with the given shift
func FirstDayOfISOWeek(weekShift int) time.Time {
	date := time.Now()
	date = time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), 0, 0, 0, date.Location())
	date = date.AddDate(0, 0, 7*weekShift)

	// iterate to Monday
	for !(date.Weekday() == time.Monday && date.Hour() == 0) {
		date = date.Add(time.Hour)
	}

	return date
}

// GetWorkRange returns a work range for a specific day
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

// GetWeekWorkRanges returns a channel of work ranges for the week
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

// ToSlice converts a channel of ranges to a slice
func ToSlice(c chan *libs.Range) []*libs.Range {
	s := make([]*libs.Range, 0)
	for r := range c {
		s = append(s, r)
	}
	return s
}
