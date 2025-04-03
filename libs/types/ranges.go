package types

import (
	"matchmaker/libs/config"
	"math/rand"
	"sort"
	"time"
)

// Range represents a time range
type Range struct {
	Start time.Time
	End   time.Time
}

// GetStart returns the start time of the range
func (r *Range) GetStart() time.Time {
	return r.Start
}

// GetEnd returns the end time of the range
func (r *Range) GetEnd() time.Time {
	return r.End
}

// Duration returns the duration of the range
func (r *Range) Duration() time.Duration {
	return r.End.Sub(r.Start)
}

// Contains returns true if the given time is within the range
func (r *Range) Contains(t time.Time) bool {
	return !t.Before(r.Start) && !t.After(r.End)
}

// Overlaps returns true if the given range overlaps with this range
func (r *Range) Overlaps(other *Range) bool {
	return !r.End.Before(other.Start) && !other.End.Before(r.Start)
}

// Before returns true if this range ends before the given time
func (r *Range) Before(t time.Time) bool {
	return r.End.Before(t)
}

// After returns true if this range starts after the given time
func (r *Range) After(t time.Time) bool {
	return r.Start.After(t)
}

func (r *Range) Pad(padding time.Duration) *Range {
	return &Range{
		Start: r.Start.Add(-padding),
		End:   r.End.Add(padding),
	}
}

// Minutes returns the duration of the range in minutes
func (r *Range) Minutes() float64 {
	return r.End.Sub(r.Start).Minutes()
}

// GenerateTimeRanges generates time ranges for the given work ranges
func GenerateTimeRanges(workRanges []*Range) []*Range {
	defaultDuration := config.GetSessionDuration()
	var durations = []time.Duration{
		defaultDuration,
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
				start = start.Add(duration)
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

// HaveIntersection checks if two ranges overlap
func HaveIntersection(range1 *Range, range2 *Range) bool {
	return range1.End.After(range2.Start) && range2.End.After(range1.Start)
}
