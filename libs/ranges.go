package libs

import (
	"github.com/spf13/viper"
	"math/rand"
	"sort"
	"time"
)

type Range struct {
	Start time.Time
	End   time.Time
}

func (r *Range) Pad(padding time.Duration) *Range {
	return &Range{
		Start: r.Start.Add(-padding),
		End:   r.End.Add(padding),
	}
}

func (r *Range) Minutes() float64 {
	return r.End.Sub(r.Start).Minutes()
}

func generateTimeRanges(workRanges []*Range) []*Range {

	defaultDuration := viper.GetDuration("sessions.sessionDurationMinutes")
	var durations = []time.Duration{
		defaultDuration * time.Minute,
	}
	ranges := []*Range{}
	for _, duration := range durations {
		for _, workRange := range workRanges {
			start := workRange.Start
			for !workRange.End.Before(start.Add(duration)) {
				ranges = append(ranges, &Range{
					Start: start,
					End:   start.Add(duration),
				})
				start = start.Add(30 * time.Minute)
			}
		}
	}

	for i := range ranges {
		j := rand.Intn(i + 1)
		ranges[i], ranges[j] = ranges[j], ranges[i]
	}

	sort.Sort(byDecreasingLength(ranges))

	return ranges
}

type byDecreasingLength []*Range

func (a byDecreasingLength) Len() int      { return len(a) }
func (a byDecreasingLength) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byDecreasingLength) Less(i, j int) bool {
	return a[j].Minutes() < a[i].Minutes()
}
